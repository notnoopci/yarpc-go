// Copyright (c) 2017 Uber Technologies, Inc.
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

package tchannel

import (
	"fmt"

	"github.com/opentracing/opentracing-go"

	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/x/config"
)

// TransportConfig configures a shared TChannel transport. This is shared
// between all TChannel outbounds and inbounds of a Dispatcher.
//
// 	transports:
// 	  tchannel:
// 	    address: :4040
type TransportConfig struct {
	// Address to listen on. Defaults to ":0" (all network interfaces and a
	// random OS-assigned port).
	Address string `config:"address,interpolate"`

	// Name of the service for TChannel. This may be omitted to use the
	// Dispatcher's service name.
	Service string `config:"service,interpolate"`
}

// InboundConfig configures a TChannel inbound.
//
// TChannel inbounds do not support any configuration parameters at this time.
type InboundConfig struct{}

// OutboundConfig configures a TChannel outbound.
//
// 	outbounds:
// 	  myservice:
// 	    tchannel:
// 	      address: 127.0.0.1:4040
type OutboundConfig struct {
	config.PeerListConfig
}

// TransportSpec returns a TransportSpec for the TChannel unary transport. See
// TransportConfig, InboundConfig, and OutboundConfig for details on the
// various supported configuration parameters.
// The given options apply BEFORE options derived from configuration, so the
// configuration takes precedence.
func TransportSpec(opts ...Option) config.TransportSpec {
	// TODO: Presets. Support "with:" and allow passing those in using
	// varargs on TransportSpec().
	var ts transportSpec
	for _, o := range opts {
		switch opt := o.(type) {
		case TransportOption:
			ts.transportOptions = append(ts.transportOptions, opt)
		// At present, TChannel inbound and outbound take no options, so no
		// further types here.
		default:
			panic(fmt.Sprintf("unknown option of type %T: %v", o, o))
		}
	}
	return ts.Spec()
}

// transportSpec holds the configurable parts of the TChannel TransportSpec.
//
// These are usually runtime dependencies that cannot be parsed from
// configuration.
type transportSpec struct {
	transportOptions []TransportOption
}

func (ts *transportSpec) Spec() config.TransportSpec {
	return config.TransportSpec{
		Name:               "tchannel",
		BuildTransport:     ts.buildTransport,
		BuildInbound:       ts.buildInbound,
		BuildUnaryOutbound: ts.buildUnaryOutbound,
	}
}

func (ts *transportSpec) buildTransport(tc *TransportConfig, k *config.Kit) (transport.Transport, error) {
	var config transportConfig
	config.tracer = opentracing.GlobalTracer()

	opts := ts.transportOptions
	for _, opt := range opts {
		opt(&config)
	}

	if tc.Address != "" {
		config.addr = tc.Address
	}
	if tc.Service != "" {
		config.name = k.ServiceName()
	}

	// Decide whether to use a channel-backed transport based on whether a
	// channel instance was provided in TransportSpec options.
	// Inbound and Outbound constructors must expect either a legacy
	// *ChannelTransport or the *Transport with peer chooser support.
	if config.ch != nil {
		return config.newChannelTransport(), nil
	}
	return config.newTransport(), nil
}

func (ts *transportSpec) buildInbound(_ *InboundConfig, t transport.Transport, k *config.Kit) (transport.Inbound, error) {
	if x, ok := t.(*ChannelTransport); ok {
		return x.NewInbound(), nil
	}
	return t.(*Transport).NewInbound(), nil
}

func (ts *transportSpec) buildUnaryOutbound(oc *OutboundConfig, t transport.Transport, k *config.Kit) (transport.UnaryOutbound, error) {
	// ChannelTransport with backing Channel support:
	if x, ok := t.(*ChannelTransport); ok {
		p, err := oc.PeerListConfig.Peer()
		if err != nil {
			return nil, fmt.Errorf(`unable to configure TChannel outbound with Channel option; TChannel ChannelTransport does not support peer lists: %v`, err)
		}
		return x.NewSingleOutbound(p), nil
	}

	// Transport with PeerChooser support:
	x := t.(*Transport)
	chooser, err := oc.PeerListConfig.BuildChooser(x, hostport.Identify, k)
	if err != nil {
		return nil, err
	}
	return x.NewOutbound(chooser), nil
}

// TODO: Document configuration parameters
