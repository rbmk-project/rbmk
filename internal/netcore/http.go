// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"

	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
)

// DialHTTP establishes a [*nop.HTTPConn].
func (nx *Network) DialHTTP(req *http.Request) (*nop.HTTPConn, error) {
	// Determine the default port
	var (
		config      = nx.NewNopConfig()
		defaultPort = ""
	)
	switch req.URL.Scheme {
	case "http":
		defaultPort = "80"
	case "https":
		defaultPort = "443"
	default:
		return nil, fmt.Errorf("unsupported scheme: %q", req.URL.Scheme)
	}

	// Determine the endpoint to connect to
	var (
		hostname = req.URL.Host
		port     = ""
	)
	if uh, up, err := net.SplitHostPort(req.URL.Host); err == nil {
		hostname, port = uh, up
	} else {
		port = defaultPort
	}
	endpoint := net.JoinHostPort(hostname, port)

	// Make the dialing pipeline
	var pipe nop.Func[netip.AddrPort, *nop.HTTPConn]
	switch req.URL.Scheme {
	case "http":
		pipe = nop.Compose2(
			nx.plainPipeline(config, "tcp"),
			nop.NewHTTPConnFuncPlain(config, nx.Logger),
		)

	case "https":
		tc := nx.TLSConfig.Clone()
		tc.ServerName = hostname
		tc.NextProtos = []string{"h2", "http/1.1"}
		pipe = nop.Compose2(
			nx.tlsPipeline(config, "tcp", tc),
			nop.NewHTTPConnFuncTLS(config, nx.Logger),
		)
	}
	runtimex.Assert(pipe != nil) // Catches refactor that breaks scheme validation

	// Defer to the common dial code
	return dial(req.Context(), nx, endpoint, pipe)
}
