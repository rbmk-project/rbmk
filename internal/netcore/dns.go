// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
)

// DNSConn is a connection to a DNS server.
type DNSConn interface {
	Exchange(ctx context.Context, query *dnscodec.Query) (*dnscodec.Response, error)
	Close() error
}

// DialDNS establishes a [*DNSConn].
func (nx *Network) DialDNS(ctx context.Context, URL *url.URL) (DNSConn, error) {
	// Determine the default port
	var (
		config      = nx.NewNopConfig()
		defaultPort = ""
	)
	switch URL.Scheme {
	case "udp", "tcp":
		defaultPort = "53"
	case "dot":
		defaultPort = "853"
	case "https":
		defaultPort = "443"
	default:
		return nil, fmt.Errorf("unsupported scheme: %q", URL.Scheme)
	}

	// Determine the endpoint to connect to
	var (
		hostname = URL.Host
		port     = ""
	)
	if uh, up, err := net.SplitHostPort(URL.Host); err == nil {
		hostname, port = uh, up
	} else {
		port = defaultPort
	}
	endpoint := net.JoinHostPort(hostname, port)

	// Make the dialing pipeline
	var pipe nop.Func[netip.AddrPort, DNSConn]
	switch URL.Scheme {
	case "udp":
		pipe = nop.Compose3(
			nx.plainPipeline(config, "udp"),
			nop.NewDNSOverUDPConnFunc(config, nx.Logger),
			dnsConnAdapter[*nop.DNSOverUDPConn]{},
		)

	case "tcp":
		pipe = nop.Compose3(
			nx.plainPipeline(config, "tcp"),
			nop.NewDNSOverTCPConnFunc(config, nx.Logger),
			dnsConnAdapter[*nop.DNSOverTCPConn]{},
		)

	case "dot":
		tc := nx.TLSConfig.Clone()
		tc.ServerName = hostname
		tc.NextProtos = []string{"dot"}
		pipe = nop.Compose3(
			nx.tlsPipeline(config, "tcp", tc),
			nop.NewDNSOverTLSConnFunc(config, nx.Logger),
			dnsConnAdapter[*nop.DNSOverTLSConn]{},
		)

	case "https":
		tc := nx.TLSConfig.Clone()
		tc.ServerName = hostname
		tc.NextProtos = []string{"h2", "http/1.1"}
		pipe = nop.Compose4(
			nx.tlsPipeline(config, "tcp", tc),
			nop.NewHTTPConnFuncTLS(config, nx.Logger),
			nop.NewDNSOverHTTPSConnFunc(config, URL.String(), nx.Logger),
			dnsConnAdapter[*nop.DNSOverHTTPSConn]{},
		)

	default:
		runtimex.Assert(false)
	}

	// Defer to the common dial code
	return dial(ctx, nx, endpoint, pipe)
}

// dnsConnAdapter adapts a compatible type to become a [DNSConn].
type dnsConnAdapter[T DNSConn] struct{}

func (dnsConnAdapter[T]) Call(ctx context.Context, conn T) (DNSConn, error) {
	return conn, nil
}
