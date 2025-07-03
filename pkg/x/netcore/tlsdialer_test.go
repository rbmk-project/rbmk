// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/rbmk-project/common/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_DialTLSContext(t *testing.T) {
	t.Run("tls config failure", func(t *testing.T) {
		nx := &Network{
			TLSConfig: nil, // Force creation of a new config
		}

		ctx := context.Background()
		conn, err := nx.DialTLSContext(ctx, "tcp", "invalid:address:format")
		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("lookup failure", func(t *testing.T) {
		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, expectedErr
			},
		}

		ctx := context.Background()
		conn, err := nx.DialTLSContext(ctx, "tcp", "example.com:443")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("dial failure", func(t *testing.T) {
		expectedErr := errors.New("mocked dial error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, expectedErr
			},
		}

		ctx := context.Background()
		conn, err := nx.DialTLSContext(ctx, "tcp", "example.com:443")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("handshake failure", func(t *testing.T) {
		mockConn := &mocks.Conn{
			MockClose: func() error {
				return nil
			},
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 443}
			},
		}

		expectedErr := errors.New("mocked handshake error")
		mockTLSConn := &mocks.TLSConn{
			Conn: mockConn,
			MockHandshakeContext: func(ctx context.Context) error {
				return expectedErr
			},
			MockConnectionState: func() tls.ConnectionState {
				return tls.ConnectionState{}
			},
		}

		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				return mockTLSConn
			},
		}

		ctx := context.Background()
		conn, err := nx.DialTLSContext(ctx, "tcp", "example.com:443")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("successful dial and handshake", func(t *testing.T) {
		mockConn := &mocks.Conn{
			MockClose: func() error {
				return nil
			},
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 443}
			},
		}

		mockTLSConn := &mocks.TLSConn{
			Conn: mockConn,
			MockHandshakeContext: func(ctx context.Context) error {
				return nil
			},
			MockConnectionState: func() tls.ConnectionState {
				return tls.ConnectionState{
					Version:            tls.VersionTLS13,
					CipherSuite:        tls.TLS_AES_128_GCM_SHA256,
					NegotiatedProtocol: "h2",
				}
			},
		}

		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				return mockTLSConn
			},
		}

		ctx := context.Background()
		conn, err := nx.DialTLSContext(ctx, "tcp", "example.com:443")
		assert.NoError(t, err)
		assert.Same(t, mockTLSConn, conn)
	})
}

func Test_tlsDialer_dial(t *testing.T) {
	t.Run("dial failure", func(t *testing.T) {
		expectedErr := errors.New("mocked dial error")

		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, expectedErr
			},
		}

		dialer := &tlsDialer{
			config: &tls.Config{},
			netx:   nx,
		}

		ctx := context.Background()
		conn, err := dialer.dial(ctx, "tcp", "example.com:443")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("handshake failure", func(t *testing.T) {
		var buf bytes.Buffer

		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}))

		expectedErr := errors.New("mocked handshake error")
		mockConn := &mocks.Conn{
			MockClose: func() error {
				return nil
			},
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 443}
			},
		}

		mockTLSConn := &mocks.TLSConn{
			Conn: mockConn,
			MockHandshakeContext: func(ctx context.Context) error {
				return expectedErr
			},
			MockConnectionState: func() tls.ConnectionState {
				return tls.ConnectionState{}
			},
		}

		nx := &Network{
			Logger: logger,
			TimeNow: func() time.Time {
				return fixedTime
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				return mockTLSConn
			},
		}

		config := &tls.Config{
			ServerName: "example.com",
		}
		dialer := &tlsDialer{
			config: config,
			netx:   nx,
		}

		ctx := context.Background()
		conn, err := dialer.dial(ctx, "tcp", "example.com:443")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)

		// We expect to see at least: connect start/done, tls start/done
		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.True(t, len(logs) >= 4, "Expected at least 4 log entries")

		// Find and verify the TLS handshake logs
		var handshakeStartFound, handshakeDoneFound bool
		for _, log := range logs {
			var logMap map[string]interface{}
			err := json.Unmarshal([]byte(log), &logMap)
			require.NoError(t, err)

			if logMap["msg"] == "tlsHandshakeStart" {
				handshakeStartFound = true
				assert.Equal(t, "127.0.0.1:1234", logMap["localAddr"])
				assert.Equal(t, "tcp", logMap["protocol"])
				assert.Equal(t, "example.com:443", logMap["remoteAddr"])
				assert.Equal(t, "unknown", logMap["tlsEngineName"])
				assert.Equal(t, "unknown", logMap["tlsParrot"])
				assert.Equal(t, "example.com", logMap["tlsServerName"])
				assert.Equal(t, false, logMap["tlsSkipVerify"])
			} else if logMap["msg"] == "tlsHandshakeDone" {
				handshakeDoneFound = true
				assert.Equal(t, expectedErr.Error(), logMap["err"])
				assert.Equal(t, "EGENERIC", logMap["errClass"])
				assert.Equal(t, "127.0.0.1:1234", logMap["localAddr"])
				assert.Equal(t, "tcp", logMap["protocol"])
				assert.Equal(t, "example.com:443", logMap["remoteAddr"])
				assert.Equal(t, "unknown", logMap["tlsEngineName"])
				assert.Equal(t, "unknown", logMap["tlsParrot"])
				assert.Equal(t, "", logMap["tlsNegotiatedProtocol"])
				assert.Equal(t, "example.com", logMap["tlsServerName"])
				assert.Equal(t, false, logMap["tlsSkipVerify"])
				assert.Equal(t, "0x0000", logMap["tlsVersion"])
			}
		}

		assert.True(t, handshakeStartFound, "tlsHandshakeStart log entry not found")
		assert.True(t, handshakeDoneFound, "tlsHandshakeDone log entry not found")
	})
}

func Test_tlsPeerCerts(t *testing.T) {
	t.Run("extracts cert from x509.HostnameError", func(t *testing.T) {
		// Create a dummy certificate
		cert := &x509.Certificate{
			Raw: []byte{1, 2, 3, 4},
		}

		// Create a hostname error with our certificate
		err := x509.HostnameError{
			Certificate: cert,
		}

		// Call the function
		certs := tlsPeerCerts(tls.ConnectionState{}, err)

		// Verify the result
		assert.Len(t, certs, 1)
		assert.Equal(t, cert.Raw, certs[0])
	})

	t.Run("extracts cert from x509.UnknownAuthorityError", func(t *testing.T) {
		// Create a dummy certificate
		cert := &x509.Certificate{
			Raw: []byte{5, 6, 7, 8},
		}

		// Create an unknown authority error with our certificate
		err := x509.UnknownAuthorityError{
			Cert: cert,
		}

		// Call the function
		certs := tlsPeerCerts(tls.ConnectionState{}, err)

		// Verify the result
		assert.Len(t, certs, 1)
		assert.Equal(t, cert.Raw, certs[0])
	})

	t.Run("extracts cert from x509.CertificateInvalidError", func(t *testing.T) {
		// Create a dummy certificate
		cert := &x509.Certificate{
			Raw: []byte{9, 10, 11, 12},
		}

		// Create a certificate invalid error with our certificate
		err := x509.CertificateInvalidError{
			Cert: cert,
		}

		// Call the function
		certs := tlsPeerCerts(tls.ConnectionState{}, err)

		// Verify the result
		assert.Len(t, certs, 1)
		assert.Equal(t, cert.Raw, certs[0])
	})

	t.Run("extracts certs from connection state", func(t *testing.T) {
		// Create some dummy certificates
		cert1 := &x509.Certificate{Raw: []byte{1, 2, 3, 4}}
		cert2 := &x509.Certificate{Raw: []byte{5, 6, 7, 8}}

		// Create a connection state with our certificates
		state := tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert1, cert2},
		}

		// Call the function with no error
		certs := tlsPeerCerts(state, nil)

		// Verify the result
		assert.Len(t, certs, 2)
		assert.Equal(t, cert1.Raw, certs[0])
		assert.Equal(t, cert2.Raw, certs[1])
	})

	t.Run("handles other errors correctly", func(t *testing.T) {
		// Create some dummy certificates in the state
		cert1 := &x509.Certificate{Raw: []byte{1, 2, 3, 4}}
		state := tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert1},
		}

		// Create a different type of error
		err := errors.New("some other error")

		// Call the function
		certs := tlsPeerCerts(state, err)

		// Verify we get certificates from the state
		assert.Len(t, certs, 1)
		assert.Equal(t, cert1.Raw, certs[0])
	})

	t.Run("returns empty slice for no certificates", func(t *testing.T) {
		// Empty state, no error
		certs := tlsPeerCerts(tls.ConnectionState{}, nil)
		assert.Empty(t, certs)
	})
}
