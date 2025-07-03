// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

// NewHTTPTransport creates an [*http.Transport] configured to use the
// given stack and the scenario's root CAs.
func (s *Scenario) NewHTTPTransport(stack *Stack) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return stack.DialContext(ctx, network, addr)
		},
		TLSClientConfig: &tls.Config{
			RootCAs: s.RootCAs(),
		},
	}
}
