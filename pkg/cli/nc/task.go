// SPDX-License-Identifier: GPL-3.0-or-later

package nc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
	"github.com/rbmk-project/rbmk/pkg/x/netcore"
)

// Task runs the `nc` task.
//
// The zero value is not ready to use. Please, make sure
// to initialize all the fields marked as MANDATORY.
type Task struct {
	// ALPNProtocols is the list of ALPN protocols to negotiate.
	ALPNProtocols []string

	// Host is the MANDATORY host to connect to.
	Host string

	// LogsWriter is the MANDATORY [io.Writer] where
	// we should write structured logs.
	LogsWriter io.Writer

	// Port is the MANDATORY port to connect to.
	Port string

	// ScanMode indicates whether we are in scan mode.
	ScanMode bool

	// ServerName is the server name to use for SNI.
	ServerName string

	// Stderr is the MANDATORY [io.Writer] for the stderr.
	Stderr io.Writer

	// Stdin is the MANDATORY [io.Reader] for the stdin.
	Stdin io.Reader

	// Stdout is the MANDATORY [io.Writer] for the stdout.
	Stdout io.Writer

	// TLSNoVerify is a flag that disables TLS verification.
	TLSNoVerify bool

	// UseTLS is a flag that ensures that we use TLS.
	UseTLS bool

	// WaitTimeout is the timeout for connect, send, and recv.
	WaitTimeout time.Duration
}

// Run runs the task and returns an error.
func (task *Task) Run(ctx context.Context) error {
	// 1. Setup logging
	logger := slog.New(slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{}))

	// 2. Create connection pool
	pool := &closepool.Pool{}
	defer pool.Close()

	// 3. Setup the network stack
	netx := &netcore.Network{}
	netx.DialContextFunc = testable.DialContext.Get()
	netx.TLSConfig = &tls.Config{
		InsecureSkipVerify: task.TLSNoVerify,
		NextProtos:         task.ALPNProtocols,
		RootCAs:            testable.RootCAs.Get(),
		ServerName:         task.ServerName,
	}
	netx.Logger = logger
	netx.WrapConn = func(ctx context.Context, netx *netcore.Network, conn net.Conn) net.Conn {
		conn = netcore.WrapConn(ctx, netx, conn)
		pool.Add(conn)
		return conn
	}

	// 5. Establish TCP and possibly TLS connection(s)
	if task.WaitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, task.WaitTimeout)
		defer cancel()
	}
	addr := net.JoinHostPort(task.Host, task.Port)
	var (
		conn net.Conn
		err  error
	)
	if task.UseTLS {
		conn, err = netx.DialTLSContext(ctx, "tcp", addr)
	} else {
		conn, err = netx.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	fmt.Fprintf(task.Stderr, "connected to %s\n", conn.RemoteAddr())

	// 6. see whether we need to route data in and out
	if !task.ScanMode {
		errc := make(chan error, 2)
		go task.copyStdinToConn(task.Stdin, conn, errc)
		go task.copyConnToStdout(conn, task.Stdout, errc)
		err = errors.Join(<-errc, <-errc)
	}

	// 7. Explicitly close the connection
	pool.Close()
	return err
}

// copyStdinToConn copies the stdin to the connection.
func (task *Task) copyStdinToConn(
	stdin io.Reader, conn net.Conn, errch chan<- error) {
	for {
		// 1. read bytes from the stdin
		const bufsiz = 4096
		buf := make([]byte, bufsiz)
		count, err := stdin.Read(buf)

		// 2. handle read error and close the write
		// side of the connection on input EOF
		if err != nil {
			if errors.Is(err, io.EOF) {
				closeWrite(conn)
				err = nil
			}
			errch <- err
			return
		}

		// 3. write bytes to the connection making sure
		// we honour the configured I/O timeout
		if task.WaitTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(task.WaitTimeout))
		}
		if _, err := conn.Write(buf[:count]); err != nil {
			errch <- err
			return
		}
		conn.SetWriteDeadline(time.Time{})
	}
}

// copyConnToStdout copies the connection to the stdout.
func (task *Task) copyConnToStdout(
	conn net.Conn, stdout io.Writer, errch chan<- error) {
	for {
		// 1. read bytes from the conn making sure
		// we honour the configured I/O timeout
		const bufsiz = 4096
		buf := make([]byte, bufsiz)
		if task.WaitTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(task.WaitTimeout))
		}
		count, err := conn.Read(buf)
		conn.SetReadDeadline(time.Time{})

		// 2. handle read error, close the stdout on EOF, and
		// always close the connection on error.
		if err != nil {
			if errors.Is(err, io.EOF) {
				maybeCloseStdout(stdout)
				err = nil
			}
			conn.Close()
			errch <- err
			return
		}

		// 3. write bytes to the stdout
		if _, err := stdout.Write(buf[:count]); err != nil {
			errch <- err
			return
		}
	}
}

// maybeCloseStdout closes the stdout if possible.
func maybeCloseStdout(stdout io.Writer) error {
	if closer, ok := stdout.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// closeWriter is an interface that allows us to close
// the write side of a connection.
type closeWriter interface {
	CloseWrite() error
}

// Ensure that [*net.TCPConn] implements [closeWriter].
var _ closeWriter = &net.TCPConn{}

// netConner is an interface that allows us to get the
// underlying [net.Conn] used by a [*tls.Conn].
type netConner interface {
	NetConn() net.Conn
}

// Ensure that [*tls.Conn] implements [netConner].
var _ netConner = &tls.Conn{}

// closeWrite closes the write side of the connection.
func closeWrite(conn net.Conn) error {
	if nc, ok := conn.(netConner); ok {
		conn = nc.NetConn()
	}
	if cw, ok := conn.(closeWriter); ok {
		return cw.CloseWrite()
	}
	return nil
}
