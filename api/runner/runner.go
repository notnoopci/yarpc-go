// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
	// IsRunning() bool
}
