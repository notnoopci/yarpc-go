package peerheap

import (
	"context"
	"fmt"
	"math"
	"sync"

	"go.uber.org/yarpc/api/peer"
	"go.uber.org/yarpc/api/transport"
	yerrors "go.uber.org/yarpc/internal/errors"
	ysync "go.uber.org/yarpc/internal/sync"
)

const (
	unavailablePenalty = math.MaxInt32
)

type List struct {
	sync.Mutex
	once ysync.LifecycleOnce

	transport peer.Transport

	byScore      peerHeap
	byIdentifier map[string]*peerScore

	added int // incrementing counter for when a peer was pushed.

	peerAvailableEvent chan struct{}
}

func (pl *List) IsRunning() bool {
	return pl.once.IsRunning()
}

func (pl *List) Start() error {
	return pl.once.Start(nil)
}

func (pl *List) Stop() error {
	return pl.once.Stop(pl.clearPeers) // TODO clear peers
}

func New(transport peer.Transport) *List {
	return &List{
		transport:          transport,
		byIdentifier:       make(map[string]*peerScore),
		peerAvailableEvent: make(chan struct{}, 1),
	}
}

func (pl *List) Update(updates peer.ListUpdates) error {
	add := updates.Additions
	remove := updates.Removals

	var errs []error

	pl.Lock()
	defer pl.Unlock()

	for _, pid := range remove {
		if err := pl.releasePeer(pid); err != nil {
			errs = append(errs, err)
		}
	}
	for _, pid := range add {
		if err := pl.retainPeer(pid); err != nil {
			errs = append(errs, err)
		}
	}

	return yerrors.MultiError(errs)
}

func (pl *List) retainPeer(pid peer.Identifier) error {
	if _, ok := pl.byIdentifier[pid.Identifier()]; ok {
		return peer.ErrPeerAddAlreadyInList(pid.Identifier())
	}

	ps := &peerScore{list: pl}
	p, err := pl.transport.RetainPeer(pid, ps)
	if err != nil {
		return err
	}

	ps.peer = p
	ps.id = pid
	ps.score = scorePeer(p)
	ps.boundFinish = ps.finish
	pl.byIdentifier[pid.Identifier()] = ps
	pl.byScore.pushPeer(ps)
	pl.internalNotifyStatusChanged(ps)
	return nil
}

func (pl *List) releasePeer(pid peer.Identifier) error {
	ps, ok := pl.byIdentifier[pid.Identifier()]
	if !ok {
		return peer.ErrPeerRemoveNotInList(pid.Identifier())
	}

	if err := pl.byScore.validate(ps); err != nil {
		return err
	}

	err := pl.transport.ReleasePeer(pid, ps)
	delete(pl.byIdentifier, pid.Identifier())
	pl.byScore.delete(ps.idx)
	ps.list = nil
	return err
}

func (pl *List) clearPeers() error {
	pl.Lock()
	defer pl.Unlock()

	var errs []error

	for {
		ps, ok := pl.byScore.peekPeer()
		if !ok {
			break
		}
		if err := pl.releasePeer(ps.id); err != nil {
			errs = append(errs, err)
		}
	}

	return yerrors.MultiError(errs)
}

func (pl *List) Choose(ctx context.Context, _ *transport.Request) (peer.Peer, func(error), error) {
	if !pl.IsRunning() {
		return nil, nil, peer.ErrPeerListNotStarted("PeerHeap")
	}

	for {
		if ps, ok := pl.get(); ok {
			pl.notifyPeerAvailable()
			return ps.peer, ps.boundFinish, nil
		}

		if err := pl.waitForPeerAvailableEvent(ctx); err != nil {
			return nil, nil, err
		}
	}
}

func (pl *List) get() (*peerScore, bool) {
	pl.Lock()
	defer pl.Unlock()

	ps, ok := pl.byScore.popPeer()
	if !ok {
		return nil, false
	}

	// Note: We push the peer back to reset the "last" timestamp.
	// This gives us round-robin behavior.
	pl.byScore.pushPeer(ps)

	return ps, ps.status.ConnectionStatus == peer.Available
}

// waitForPeerAvailableEvent waits until a peer is added to the peer list or the
// given context finishes.
// Must NOT be run in a mutex.Lock()
func (pl *List) waitForPeerAvailableEvent(ctx context.Context) error {
	if _, ok := ctx.Deadline(); !ok {
		return peer.ErrChooseContextHasNoDeadline("PeerHeap")
	}

	select {
	case <-pl.peerAvailableEvent:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// notifyPeerAvailable writes to a channel indicating that a Peer is currently
// available for requests
func (pl *List) notifyPeerAvailable() {
	select {
	case pl.peerAvailableEvent <- struct{}{}:
	default:
	}
}

func (pl *List) NotifyStatusChanged(pid peer.Identifier) {
	pl.Lock()
	ps := pl.byIdentifier[pid.Identifier()]
	pl.Unlock()
	ps.NotifyStatusChanged(pid)
}

func (pl *List) notifyStatusChanged(ps *peerScore) {
	pl.Lock()
	p := ps.peer
	ps.status = p.Status()
	ps.score = scorePeer(p)
	pl.byScore.update(ps.idx)
	pl.Unlock()

	if p.Status().ConnectionStatus == peer.Available {
		pl.notifyPeerAvailable()
	}
}

func (pl *List) internalNotifyStatusChanged(ps *peerScore) {
	p := ps.peer
	ps.status = p.Status()
	ps.score = scorePeer(p)
	pl.byScore.update(ps.idx)

	if p.Status().ConnectionStatus == peer.Available {
		pl.notifyPeerAvailable()
	}
}

func scorePeer(p peer.Peer) int {
	status := p.Status()
	score := status.PendingRequestCount
	if status.ConnectionStatus != peer.Available {
		score += unavailablePenalty
	}
	return score
}

func (pl *List) String() string {
	return fmt.Sprintf("{len:%d}", pl.byScore.Len())
}
