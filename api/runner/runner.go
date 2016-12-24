package runner

import "context"

// Runner exposes an interface for runnables, including transports, inbounds, and outbounds.
type Runner interface {
	// Run takes a cancellable context, starts the runner, runs until the
	// context is cancelled, drains any outstanding work, then stops.
	// A runner can only run once.
	// Subsequent calls to run are invalid.
	// The run function closes a "started" channel when it is ready to accept
	// work, and closes a "stopped" channel when it has gracefully concluded
	// all work.
	Run(ctx context.Context) error

	// Started returns a channel that will close when the runner has started and is
	// able to accept work.
	Started() <-chan struct{}

	// Stopped returns a channel that will close when the runner has stopped
	// and all work has drained.
	Stopped() <-chan struct{}

	// IsRunning returns whether the runner has started and hasn't yet stopped.
	IsRunning() bool
}
