//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/measurexlite/conn.go
//
// Conn wrapper.
//

package netcore

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/rbmk-project/common/errclass"
)

// connLocalAddr is a safe way to get the local address of a connection.
func connLocalAddr(conn net.Conn) net.Addr {
	if conn != nil && conn.LocalAddr() != nil {
		return conn.LocalAddr()
	}
	return emptyAddr{}
}

// connRemoteAddr is a safe way to get the remote address of a connection.
func connRemoteAddr(conn net.Conn) net.Addr {
	if conn != nil && conn.RemoteAddr() != nil {
		return conn.RemoteAddr()
	}
	return emptyAddr{}
}

// emptyAddr is an empty [net.Addr].
type emptyAddr struct{}

// Network implements [net.Addr].
func (emptyAddr) Network() string { return "" }

// String implements [net.Addr].
func (emptyAddr) String() string { return "" }

// maybeWrapConn wraps a connection when it makes sense to do so.
func (nx *Network) maybeWrapConn(ctx context.Context, conn net.Conn) net.Conn {
	if conn != nil && nx.Logger != nil && nx.WrapConn != nil {
		conn = nx.WrapConn(ctx, nx, conn)
	}
	return conn
}

// WrapConn wraps a given [net.Conn] to emit structured logs.
//
// The context argument is only used for logging and does not constrain
// in any way the lifetime of the wrapped connection.
func WrapConn(ctx context.Context, netx *Network, conn net.Conn) net.Conn {
	laddr := connLocalAddr(conn)
	conn = &connWrapper{
		ctx:       ctx,
		closeonce: sync.Once{},
		conn:      conn,
		laddr:     laddr.String(),
		netx:      netx,
		protocol:  laddr.Network(),
		raddr:     connRemoteAddr(conn).String(),
	}
	return conn
}

// connWrapper wraps a [net.Conn].
type connWrapper struct {
	ctx       context.Context // only used for logging
	closeonce sync.Once
	conn      net.Conn
	laddr     string
	netx      *Network // may contain nil logger!
	protocol  string
	raddr     string
}

// Close implements [net.Conn].
func (c *connWrapper) Close() (err error) {
	c.closeonce.Do(func() {
		t0 := c.netx.timeNow()
		if c.netx.Logger != nil {
			c.netx.Logger.InfoContext(
				c.ctx,
				"closeStart",
				slog.String("localAddr", c.laddr),
				slog.String("protocol", c.protocol),
				slog.String("remoteAddr", c.raddr),
				slog.Time("t", t0),
			)
		}

		err = c.conn.Close()

		if c.netx.Logger != nil {
			c.netx.Logger.InfoContext(
				c.ctx,
				"closeDone",
				slog.Any("err", err),
				slog.String("errClass", errclass.New(err)),
				slog.String("localAddr", c.laddr),
				slog.String("protocol", c.protocol),
				slog.String("remoteAddr", c.raddr),
				slog.Time("t0", t0),
				slog.Time("t", c.netx.timeNow()),
			)
		}
	})
	return
}

// LocalAddr implements [net.Conn].
func (c *connWrapper) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Read implements [net.Conn].
func (c *connWrapper) Read(buf []byte) (int, error) {
	t0 := c.netx.timeNow()
	if c.netx.Logger != nil {
		c.netx.Logger.InfoContext(
			c.ctx,
			"readStart",
			slog.Int("ioBufferSize", len(buf)),
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t", t0),
		)
	}

	count, err := c.conn.Read(buf)

	if c.netx.Logger != nil {
		c.netx.Logger.InfoContext(
			c.ctx,
			"readDone",
			slog.Int("ioBytesCount", count),
			slog.Any("err", err),
			slog.String("errClass", errclass.New(err)),
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t0", t0),
			slog.Time("t", c.netx.timeNow()),
		)
	}

	return count, err
}

// RemoteAddr implements [net.Conn].
func (c *connWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements [net.Conn].
func (c *connWrapper) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline implements [net.Conn].
func (c *connWrapper) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements [net.Conn].
func (c *connWrapper) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// Write implements [net.Conn].
func (c *connWrapper) Write(data []byte) (n int, err error) {
	t0 := c.netx.timeNow()
	if c.netx.Logger != nil {
		c.netx.Logger.InfoContext(
			c.ctx,
			"writeStart",
			slog.Int("ioBufferSize", len(data)),
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t", t0),
		)
	}

	count, err := c.conn.Write(data)

	if c.netx.Logger != nil {
		c.netx.Logger.InfoContext(
			c.ctx,
			"writeDone",
			slog.Int("ioBytesCount", count),
			slog.Any("err", err),
			slog.String("errClass", errclass.New(err)),
			slog.String("localAddr", c.laddr),
			slog.String("protocol", c.protocol),
			slog.String("remoteAddr", c.raddr),
			slog.Time("t0", t0),
			slog.Time("t", c.netx.timeNow()),
		)
	}

	return count, err
}
