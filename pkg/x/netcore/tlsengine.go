//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/tls.go
//

package netcore

import (
	"crypto/tls"
	"net"
)

// TLSEngine provides methods to create and describe [TLSConn] instances.
type TLSEngine interface {
	// Name returns the TLS engine name (e.g., "stdlib", "utls").
	Name() string

	// NewClientConn creates a new [TLSConn] instance.
	NewClientConn(conn net.Conn, config *tls.Config) TLSConn

	// Parrot returns the fingerprint we're parroting (e.g., "chrome",
	// "firefox") or an empty string if there's no parroting.
	Parrot() string
}

// newTLSEngine creates a new [TLSEngine] instance.
func (nx *Network) newTLSEngine() TLSEngine {
	switch {
	case nx.TLSEngine != nil:
		return nx.TLSEngine
	case nx.NewTLSClientConn != nil:
		return tlsEngineUnknown(nx.NewTLSClientConn)
	default:
		return &TLSEngineStdlib{}
	}
}

// TLSEngineStdlib is a [TLSEngine] using the Go standard library.
type TLSEngineStdlib struct{}

// Ensure that [*TLSEngineStdlib] implements [TLSEngine].
var _ TLSEngine = &TLSEngineStdlib{}

// Name implements [TLSEngine] and returns "stdlib".
func (*TLSEngineStdlib) Name() string {
	return "stdlib"
}

// NewClientConn implements [TLSEngine] and uses the standard
// library [tls.Client] function to create a [TLSConn].
func (*TLSEngineStdlib) NewClientConn(conn net.Conn, config *tls.Config) TLSConn {
	return tls.Client(conn, config)
}

// Parrot implements [TLSEngine] and returns an empty string.
func (*TLSEngineStdlib) Parrot() string {
	return ""
}

// tlsEngineUnknown is an adapter that turns a legacy NewTLSClientConn function
// into a [TLSEngine] implementation with "unknown" metadata values.
type tlsEngineUnknown func(conn net.Conn, config *tls.Config) TLSConn

// Ensure that [tlsEngineUnknown] implements [TLSEngine].
var _ TLSEngine = tlsEngineUnknown(nil)

// Name implements [TLSEngine].
func (t tlsEngineUnknown) Name() string {
	return "unknown"
}

// NewClientConn implements [TLSEngine].
func (t tlsEngineUnknown) NewClientConn(conn net.Conn, config *tls.Config) TLSConn {
	return t(conn, config)
}

// Parrot implements [TLSEngine].
func (t tlsEngineUnknown) Parrot() string {
	return "unknown"
}
