// SPDX-License-Identifier: GPL-3.0-or-later

// Package netcore is RBMK's core networking library.
package netcore

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
	"github.com/rbmk-project/rbmk/internal/testablenet"
	"github.com/rbmk-project/rbmk/pkg/common/errclass"
)

// Resolver is a [*net.Resolver]-like abstraction.
type Resolver interface {
	LookupHost(ctx context.Context, domain string) ([]string, error)
}

// DialContextFunc is the function for creating new [net.Conn] instances.
type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

type dialerAdapter struct {
	fx DialContextFunc
}

var _ nop.Dialer = dialerAdapter{}

// DialContext implements [nop.Dialer].
func (d dialerAdapter) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return d.fx(ctx, network, address)
}

// resolverAdapter adapts a [testablenet.LookupHostFunc] to the [Resolver] interface.
type resolverAdapter struct {
	fx testablenet.LookupHostFunc
}

var _ Resolver = resolverAdapter{}

// LookupHost implements [Resolver].
func (r resolverAdapter) LookupHost(ctx context.Context, domain string) ([]string, error) {
	return r.fx(ctx, domain)
}

// Network allows to create network connections.
//
// Use [NewNetwork] to construct.
type Network struct {
	// DialContextFunc is the function for creating a new conn.
	//
	// The [NewNetwork] function initializes this using [testablenet.DialContext].
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// Logger is the logger to use.
	//
	// The [NewNetwork] function initializes this using JSON slogger writing on the [os.Stderr].
	Logger *slog.Logger

	// Resolver is the resolver to use.
	//
	// The [NewNetwork] function initializes this using [testablenet.LookupHost].
	Resolver Resolver

	// SplitHostPort is the function that splits the endpoint to resolve into
	// a hostname (or IPv4/IPv6 address) and a port.
	//
	// The [NewNetwork] function initializes this to [net.SplitHostPort].
	SplitHostPort func(endpoint string) (hostname string, port string, err error)

	// TLSConfig is the TLS config to use.
	//
	// The [NewNetwork] function initializes this using a [*tls.Config]
	// with the root CAs from [testablenet.RootCAs].
	TLSConfig *tls.Config

	// TimeNow is the function to get the current time.
	//
	// The [NewNetwork] function initializes this using [time.Now].
	TimeNow func() time.Time
}

// NewNetwork creates a new [*Network] with default values.
func NewNetwork() *Network {
	return &Network{
		DialContextFunc: testablenet.DialContext.Get(),
		Logger:          slog.New(slog.NewJSONHandler(os.Stderr, nil)),
		Resolver:        resolverAdapter{testablenet.LookupHost.Get()},
		SplitHostPort:   net.SplitHostPort,
		TLSConfig:       &tls.Config{RootCAs: testablenet.RootCAs.Get()},
		TimeNow:         time.Now,
	}
}

// NewNopConfig creates a new [*nop.Config] instance.
func (nx *Network) NewNopConfig() *nop.Config {
	return &nop.Config{
		Dialer:        dialerAdapter{nx.DialContextFunc},
		ErrClassifier: nop.ErrClassifierFunc(errclass.New),
		TimeNow:       nx.TimeNow,
	}
}

// DialContext establishes a new TCP/UDP [net.Conn].
func (nx *Network) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return dial(ctx, nx, address, nx.plainPipeline(nx.NewNopConfig(), network))
}

func (nx *Network) plainPipeline(config *nop.Config, network string) nop.Func[netip.AddrPort, net.Conn] {
	return nop.Compose3(
		nop.NewConnectFunc(config, network, nx.Logger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(config, nx.Logger),
	)
}

// DialTLSContext establishes a new TLS [net.Conn].
//
// The caller must set [Network.TLSConfig].ServerName before calling
// this method. Unlike [Network.DialDNS] and [Network.DialHTTP], this
// method does not automatically derive the ServerName from the address.
func (nx *Network) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	return dial(ctx, nx, address, nop.Compose2(
		nx.tlsPipeline(nx.NewNopConfig(), network, nx.TLSConfig.Clone()),
		tlsConnAdapter{},
	))
}

func (nx *Network) tlsPipeline(cfg *nop.Config, network string, tcfg *tls.Config) nop.Func[netip.AddrPort, nop.TLSConn] {
	return nop.Compose4(
		nop.NewConnectFunc(cfg, network, nx.Logger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(cfg, nx.Logger),
		nop.NewTLSHandshakeFunc(cfg, tcfg, nx.Logger),
	)
}

// pipeline is the generic pipeline to create a new [net.Conn].
type pipeline[T any] nop.Func[netip.AddrPort, T]

// tlsConnAdapter adapts [nop.TLSConn] to be a [net.Conn].
type tlsConnAdapter struct{}

// Call implements [nop.Func].
func (tlsConnAdapter) Call(ctx context.Context, conn nop.TLSConn) (net.Conn, error) {
	return conn, nil
}

// dial is the internal function used for dialing.
func dial[T any](ctx context.Context, nx *Network, address string, pipe pipeline[T]) (T, error) {
	// Create zero value for error returns
	var zero T

	// Unpack the network endpoint
	domain, port, err := nx.SplitHostPort(address)
	if err != nil {
		return zero, err
	}

	// Map the domain to addresses
	addrs, err := nx.Resolver.LookupHost(ctx, domain)
	if err != nil {
		return zero, err
	}
	runtimex.Assert(len(addrs) >= 1)

	// Attempt dialing with each address
	var errv []error
	for _, addr := range addrs {
		epnt, err := netip.ParseAddrPort(net.JoinHostPort(addr, port))
		if err != nil {
			errv = append(errv, err)
			continue
		}
		conn, err := pipe.Call(ctx, epnt)
		if err != nil {
			errv = append(errv, err)
			continue
		}
		return conn, nil
	}
	return zero, errors.Join(errv...)
}
