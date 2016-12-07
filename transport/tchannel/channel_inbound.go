package tchannel

import (
	"github.com/opentracing/opentracing-go"
	"go.uber.org/yarpc/internal/errors"
	"go.uber.org/yarpc/transport"
)

type inboundConfig struct {
}

// ChannelInbound is a TChannel Inbound backed by a pre-existing TChannel
// Channel.
type ChannelInbound struct {
	ch        Channel
	addr      string
	registry  transport.Registry
	tracer    opentracing.Tracer
	transport *ChannelTransport
}

// NewInbound returns a new TChannel inbound backed by a shared TChannel
// transport.  The returned ChannelInbound does not support peer.Chooser
// and uses TChannel's own internal load balancing peer selection.
func (t *ChannelTransport) NewInbound() *ChannelInbound {
	return &ChannelInbound{
		ch:        t.ch,
		tracer:    t.tracer,
		transport: t,
	}
}

// SetRegistry configures a registry to handle incoming requests.
// This satisfies the transport.Inbound interface, and would be called
// by a dispatcher when it starts.
func (i *ChannelInbound) SetRegistry(registry transport.Registry) {
	i.registry = registry
}

// Transports returns a slice containing the ChannelInbound's underlying
// ChannelTransport.
func (i *ChannelInbound) Transports() []transport.Transport {
	return []transport.Transport{i.transport}
}

// Channel returns the underlying Channel for this Inbound.
func (i *ChannelInbound) Channel() Channel {
	return i.ch
}

func (i *ChannelInbound) Start() error {
	if i.registry == nil {
		return errors.NoRegistryError{}
	}

	// Set up handlers. This must occur after construction because the
	// dispatcher, or its equivalent, calls SetRegistry before Start.
	// This also means that starting inbounds should block starting the transport.
	sc := i.ch.GetSubChannel(i.ch.ServiceName())
	existing := sc.GetHandlers()
	sc.SetHandler(handler{existing: existing, Registry: i.registry, tracer: i.tracer})

	return nil
}

func (i *ChannelInbound) Stop() error {
	return nil
}
