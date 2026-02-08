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

	// TLSConfig is the TLS config to use.
	//
	// The [NewNetwork] function initializes this using a [*tls.Config]
	// with the root CAs from [testablenet.RootCAs].
	TLSConfig *tls.Config

	// Resolver is the resolver to use.
	//
	// The [NewNetwork] function initializes this using an zero-initialized [*net.Resolver].
	Resolver Resolver

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
		TLSConfig:       &tls.Config{RootCAs: testablenet.RootCAs.Get()},
		Resolver:        &net.Resolver{},
		TimeNow:         time.Now,
	}
}

// newNopConfig creates a new [*nop.Config] instance.
func (nx *Network) newNopConfig() *nop.Config {
	return &nop.Config{
		Dialer:        dialerAdapter{nx.DialContextFunc},
		ErrClassifier: nop.ErrClassifierFunc(errclass.New),
		TimeNow:       nx.TimeNow,
	}
}

// DialContext establishes a new TCP/UDP [net.Conn].
func (nx *Network) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	config := nx.newNopConfig()
	return nx.dial(ctx, address, nop.Compose3(
		nop.NewConnectFunc(config, network, nx.Logger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(config, nx.Logger),
	))
}

// DialTLSContext establishes a new TLS [net.Conn].
func (nx *Network) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	config := nx.newNopConfig()
	return nx.dial(ctx, address, nop.Compose5(
		nop.NewConnectFunc(config, network, nx.Logger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(config, nx.Logger),
		nop.NewTLSHandshakeFunc(config, nx.TLSConfig, nx.Logger),
		tlsConnAdapter{},
	))
}

// pipeline is the generic pipeline to create a new [net.Conn].
type pipeline nop.Func[netip.AddrPort, net.Conn]

// tlsConnAdapter adapts [nop.TLSConn] to be a [net.Conn].
type tlsConnAdapter struct{}

// Call implements [nop.Func].
func (tlsConnAdapter) Call(ctx context.Context, conn nop.TLSConn) (net.Conn, error) {
	return conn, nil
}

// dial is the internal function used for dialing.
func (nx *Network) dial(ctx context.Context, address string, pipe pipeline) (net.Conn, error) {
	// Unpack the network endpoint
	domain, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	// Map the domain to addresses
	addrs, err := nx.Resolver.LookupHost(ctx, domain)
	if err != nil {
		return nil, err
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
	return nil, errors.Join(errv...)
}
