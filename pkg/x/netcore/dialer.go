//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
//
// Cleartext conn dialer.
//

package netcore

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/rbmk-project/common/errclass"
)

// DialContext establishes a new TCP/UDP connection.
func (nx *Network) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// resolve the endpoints to connect to
	endpoints, err := nx.maybeLookupEndpoint(ctx, address)
	if err != nil {
		return nil, err
	}

	// sequentially attempt with each available endpoint
	return nx.sequentialDial(ctx, network, nx.dialLog, endpoints...)
}

// dialContextFunc is a function used to dial a connection.
type dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

// sequentialDial attempts to dial the endpoints in sequence until one
// of them succeeds. It returns the first successfully established network
// connection, on success, and the union of all errors, otherwise.
func (nx *Network) sequentialDial(
	ctx context.Context,
	network string,
	fx dialContextFunc,
	endpoints ...string,
) (net.Conn, error) {
	// TODO(bassosimone): decide whether we want to sort IPv4 before IPv6
	// here, and whether we want another method for happy eyeballs.
	var errv []error
	for _, endpoint := range endpoints {
		conn, err := fx(ctx, network, endpoint)
		if conn != nil && err == nil {
			return conn, nil
		}
		errv = append(errv, err)
	}
	if len(errv) <= 0 {
		return nil, errors.New("no endpoints to dial")
	}
	return nil, errors.Join(errv...)
}

// dialLog dials and emits structured logs.
func (nx *Network) dialLog(ctx context.Context, network, address string) (net.Conn, error) {
	// Optionally enforce timeout for connection establishment
	if nx.DialContextTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, nx.DialContextTimeout)
		defer cancel()
	}

	// Emit structured event before the dial
	t0 := nx.emitConnectStart(ctx, network, address)

	// Establish the connection
	conn, err := nx.dialNet(ctx, network, address)

	// Emit structured event after the dial
	nx.emitConnectDone(ctx, network, address, t0, conn, err)

	// Maybe wrap the connection if it's not nil and it makes sense
	// to wrap it (i.e., we have logging enabled)
	conn = nx.maybeWrapConn(ctx, conn)

	// Return the connection and error to the caller
	return conn, err
}

// defaultDialer is the default [*net.Dialer] we use.
var defaultDialer = func() *net.Dialer {
	dialer := &net.Dialer{}
	dialer.SetMultipathTCP(false)
	return dialer
}()

// dialNet dials using the net package or the configured dialing override.
func (nx *Network) dialNet(ctx context.Context, network, address string) (net.Conn, error) {
	// if there's an user provided dialer func, use it
	if nx.DialContextFunc != nil {
		return nx.DialContextFunc(ctx, network, address)
	}

	// otherwise fallback to a default dialer
	child := defaultDialer
	if nx.NewDialerOrSingleton != nil {
		child = nx.NewDialerOrSingleton()
	}
	return child.DialContext(ctx, network, address)
}

// emitConnectStart emits a structured event before the dial.
func (nx *Network) emitConnectStart(ctx context.Context, network, address string) time.Time {
	t0 := nx.timeNow()
	if nx.Logger != nil {
		nx.Logger.InfoContext(
			ctx,
			"connectStart",
			slog.String("protocol", network),
			slog.String("remoteAddr", address),
			slog.Time("t", t0),
		)
	}
	return t0
}

// emitConnectDone emits a structured event after the dial.
func (nx *Network) emitConnectDone(ctx context.Context,
	network, address string, t0 time.Time, conn net.Conn, err error) {
	if nx.Logger != nil {
		nx.Logger.InfoContext(
			ctx,
			"connectDone",
			slog.Any("err", err),
			slog.String("errClass", errclass.New(err)),
			slog.String("localAddr", connLocalAddr(conn).String()),
			slog.String("protocol", network),
			slog.String("remoteAddr", address),
			slog.Time("t0", t0),
			slog.Time("t", nx.timeNow()),
		)
	}
}
