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

package cherami

import (
	"testing"

	"go.uber.org/yarpc/transport/x/cherami/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOutbound(t *testing.T) {
	mockPublisher := &mocks.Publisher{}
	mockPublisher.On(`Close`)
	mockFactory := &mocks.ClientFactory{}
	mockFactory.On(`GetPublisher`, nil, mock.Anything, mock.Anything).Return(mockPublisher, nil, nil)
	transport := NewTransport(nil)
	outbound := transport.NewOutbound(OutboundConfig{
		Destination: `dest`,
	})
	outbound.setClientFactory(mockFactory)
	err := outbound.Start()
	assert.Nil(t, err)

	err = outbound.Stop()
	assert.Nil(t, err)
}
