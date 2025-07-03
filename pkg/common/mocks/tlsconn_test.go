// SPDX-License-Identifier: GPL-3.0-or-later

package mocks

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestTLSConn(t *testing.T) {
	t.Run("ConnectionState", func(t *testing.T) {
		expectedState := tls.ConnectionState{
			Version:                     tls.VersionTLS13,
			HandshakeComplete:           true,
			DidResume:                   false,
			CipherSuite:                 tls.TLS_AES_128_GCM_SHA256,
			NegotiatedProtocol:          "h2",
			ServerName:                  "example.com",
			PeerCertificates:            nil,
			VerifiedChains:              nil,
			SignedCertificateTimestamps: nil,
			OCSPResponse:                nil,
		}

		conn := &TLSConn{
			MockConnectionState: func() tls.ConnectionState {
				return expectedState
			},
		}

		state := conn.ConnectionState()
		if diff := cmp.Diff(expectedState, state,
			cmpopts.IgnoreUnexported(tls.ConnectionState{})); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("HandshakeContext", func(t *testing.T) {
		expected := errors.New("mocked handshake error")
		conn := &TLSConn{
			MockHandshakeContext: func(ctx context.Context) error {
				return expected
			},
		}

		err := conn.HandshakeContext(context.Background())
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
	})

	t.Run("Embedded Conn methods", func(t *testing.T) {
		expected := errors.New("mocked read error")
		conn := &TLSConn{
			Conn: &Conn{
				MockRead: func(b []byte) (int, error) {
					return 0, expected
				},
			},
		}

		count, err := conn.Read(make([]byte, 128))
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
		if count != 0 {
			t.Fatal("expected 0 bytes")
		}
	})
}
