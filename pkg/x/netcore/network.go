//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Definition of Network.
//

package netcore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"net"
	"time"
)

// Network allows dialing and measuring TCP/UDP/TLS connections.
//
// The zero value is ready to use.
//
// A [*Network] is safe for concurrent use by multiple goroutines as long as
// you don't modify its fields after construction and the underlying fields you
// may set (e.g., DialContextFunc) are also safe.
type Network struct {
	// DialContextFunc is the optional dialer for creating new
	// TCP and UDP connections. If this field is nil, the default
	// dialer from the [net] package will be used.
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// Logger is the optional structured logger for emitting
	// structured diagnostic events. If this field is nil, we
	// will not be emitting structured logs.
	Logger *slog.Logger

	// LookupHostFunc is the optional function to resolve a domain
	// name to IP addresses. If this field is nil, we use the
	// default [*net.Resolver] from the [net] package.
	LookupHostFunc func(ctx context.Context, domain string) ([]string, error)

	// NewTLSClientConn is the optional function to create a new TLS client
	// connection. If this field is nil, we use the [crypto/tls] package.
	//
	// If this field is not nil and TLSEngine is also not nil, the latter
	// will take precedence in creating a new [TLSConn].
	//
	// When using this field, the "tlsEngineName" and "tlsParrot" fields
	// in the structured logs will both be set to "unknown".
	//
	// Deprecated: use the TLSEngine field instead.
	NewTLSClientConn func(conn net.Conn, config *tls.Config) TLSConn

	// RootCAs contains the optional [*x509.CertPool] used when
	// creating TLS connections. If it is not set, we use the system's
	// root CAs. This field is only used when the TLSConfig field is nil.
	RootCAs *x509.CertPool

	// TLSConfig is the TLS client config to use. If this field is nil, we
	// will try to create a suitable config based on the network and address
	// that are passed to the DialTLSContext method.
	TLSConfig *tls.Config

	// TimeNow is an optional function that returns the current time.
	// If this field is nil, the [time.Now] function will be used.
	TimeNow func() time.Time

	// WrapConn is an optional function to wrap a connection to emit
	// structured logs. [WrapConn] is the default wrapper to use.
	WrapConn func(ctx context.Context, netx *Network, conn net.Conn) net.Conn

	// LookupHostTimeout is the optional timeout to use for limiting
	// the maximum time spent resolving a domain name.
	LookupHostTimeout time.Duration

	// DialContextTimeout is the optional timeout to use for limiting
	// the maximum time spent creating a single connection.
	DialContextTimeout time.Duration

	// NewResolverOrSingleton is the optional function that returns
	// the [*net.Resolver] to use when LookupHostFunc is not set. As the
	// name suggests, this function may either create a new [*net.Resolver]
	// for each call or just return a singleton instance. When this method
	// is not set, we use an internal zero-initialized, static [*net.Resolver].
	NewResolverOrSingleton func() *net.Resolver

	// NewDialerOrSingleton is the optional function that returns
	// the [*net.Dialer] to use when DialContextFunc is not set. As the
	// name suggests, this function may either create a new [*net.Dialer]
	// for each call or just return a singleton instance. When this method
	// is not set, we use an internal, static [*net.Dialer] where
	// support for Multipath TCP has been disabled. We disable Multipath
	// TCP because we focus on precise internet measurements.
	NewDialerOrSingleton func() *net.Dialer

	// TLSEngine is the optional [TLSEngine] to use for creating a new
	// instance of [TLSConn]. If this field is nil, we create on the fly
	// and use an instance of [TLSEngineStdlib].
	TLSEngine TLSEngine
}

// DefaultNetwork is the default [*Network] used by this package.
var DefaultNetwork = &Network{}

// timeNow is a function that returns the current time.
func (nx *Network) timeNow() time.Time {
	if nx.TimeNow != nil {
		return nx.TimeNow()
	}
	return time.Now()
}
