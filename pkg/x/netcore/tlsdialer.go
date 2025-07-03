//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/measurexlite/tls.go
//
// TLS dialing code
//

package netcore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/rbmk-project/common/errclass"
)

// TLSConn is the interface implementing [*tls.Conn] as well as
// the conn exported by alternative TLS libraries.
type TLSConn interface {
	ConnectionState() tls.ConnectionState
	HandshakeContext(ctx context.Context) error
	net.Conn
}

// DialTLSContext establishes a new TLS connection.
func (nx *Network) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	// obtain the TLS config to use
	config, err := nx.tlsConfig(network, address)
	if err != nil {
		return nil, err
	}

	// resolve the endpoints to connect to
	endpoints, err := nx.maybeLookupEndpoint(ctx, address)
	if err != nil {
		return nil, err
	}

	// build a TLS dialer
	td := &tlsDialer{config: config, netx: nx}

	// sequentially attempt with each available endpoint
	return nx.sequentialDial(ctx, network, td.dial, endpoints...)
}

type tlsDialer struct {
	config *tls.Config
	netx   *Network
}

func (td *tlsDialer) dial(ctx context.Context, network, address string) (net.Conn, error) {
	// dial and log the results of dialing
	conn, err := td.netx.dialLog(ctx, network, address)
	if err != nil {
		return nil, err
	}

	// create TLS client connection
	engine := td.netx.newTLSEngine()
	tconn := engine.NewClientConn(conn, td.config)

	// emit event before the TLS handshake
	laddr := connLocalAddr(conn).String()
	t0 := td.emitTLSHandshakeStart(ctx, laddr, network, address, engine)

	// perform the TLS handshake
	err = tconn.HandshakeContext(ctx)

	// emit event after the TLS handshake
	td.emitTLSHandshakeDone(
		ctx, laddr, network, address, engine, t0, err, tconn.ConnectionState())

	// process the results
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tconn, nil
}

// emitTLSHandshakeStart emits a TLS handshake start event.
func (td *tlsDialer) emitTLSHandshakeStart(ctx context.Context,
	localAddr, network, remoteAddr string, engine TLSEngine) time.Time {
	t0 := td.netx.timeNow()
	if td.netx.Logger != nil {
		td.netx.Logger.InfoContext(
			ctx,
			"tlsHandshakeStart",
			slog.String("localAddr", localAddr),
			slog.String("protocol", network),
			slog.String("remoteAddr", remoteAddr),
			slog.Time("t", t0),
			slog.String("tlsEngineName", engine.Name()),
			slog.String("tlsParrot", engine.Parrot()),
			slog.String("tlsServerName", td.config.ServerName),
			slog.Bool("tlsSkipVerify", td.config.InsecureSkipVerify),
		)
	}
	return t0
}

// emitTLSHandshakeDone emits a TLS handshake done event.
func (td *tlsDialer) emitTLSHandshakeDone(ctx context.Context,
	localAddr, network, remoteAddr string, engine TLSEngine,
	t0 time.Time, err error, state tls.ConnectionState) {
	if td.netx.Logger != nil {
		td.netx.Logger.InfoContext(
			ctx,
			"tlsHandshakeDone",
			slog.Any("err", err),
			slog.String("errClass", errclass.New(err)),
			slog.String("localAddr", localAddr),
			slog.String("protocol", network),
			slog.String("remoteAddr", remoteAddr),
			slog.Time("t0", t0),
			slog.Time("t", td.netx.timeNow()),
			slog.String("tlsCipherSuite", tls.CipherSuiteName(state.CipherSuite)),
			slog.String("tlsEngineName", engine.Name()),
			slog.String("tlsParrot", engine.Parrot()),
			slog.String("tlsNegotiatedProtocol", state.NegotiatedProtocol),
			slog.Any("tlsPeerCerts", tlsPeerCerts(state, err)),
			slog.String("tlsServerName", td.config.ServerName),
			slog.Bool("tlsSkipVerify", td.config.InsecureSkipVerify),
			slog.String("tlsVersion", tls.VersionName(state.Version)),
		)
	}
}

// tlsPeerCerts extracts the certificates either from the list of certificates
// in the connection state or from the error that occurred.
func tlsPeerCerts(
	state tls.ConnectionState, err error) (out [][]byte) {
	out = [][]byte{}

	// 1. Check whether the error is a known certificate error and extract
	// the certificate using `errors.As` for additional robustness.
	var x509HostnameError x509.HostnameError
	if errors.As(err, &x509HostnameError) {
		// Test case: https://wrong.host.badssl.com/
		out = append(out, x509HostnameError.Certificate.Raw)
		return
	}

	var x509UnknownAuthorityError x509.UnknownAuthorityError
	if errors.As(err, &x509UnknownAuthorityError) {
		// Test case: https://self-signed.badssl.com/. This error has
		// never been among the ones returned by MK.
		out = append(out, x509UnknownAuthorityError.Cert.Raw)
		return
	}

	var x509CertificateInvalidError x509.CertificateInvalidError
	if errors.As(err, &x509CertificateInvalidError) {
		// Test case: https://expired.badssl.com/
		out = append(out, x509CertificateInvalidError.Cert.Raw)
		return
	}

	// 2. Otherwise extract certificates from the connection state.
	for _, cert := range state.PeerCertificates {
		out = append(out, cert.Raw)
	}
	return
}
