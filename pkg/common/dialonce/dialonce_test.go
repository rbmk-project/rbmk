// SPDX-License-Identifier: GPL-3.0-or-later

package dialonce

import (
	"context"
	"errors"
	"net"
	"testing"
)

// mockConn is a minimal implementation of net.Conn for testing
type mockConn struct {
	net.Conn // embedded to provide default implementations
}

func TestWrap(t *testing.T) {
	t.Run("successful single dial", func(t *testing.T) {
		dialCount := 0
		mockDial := func(ctx context.Context, network, address string) (net.Conn, error) {
			dialCount++
			return &mockConn{}, nil
		}

		wrapped := Wrap(mockDial)
		conn, err := wrapped(context.Background(), "tcp", "example.com:80")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if conn == nil {
			t.Error("expected connection, got nil")
		}
		if dialCount != 1 {
			t.Errorf("expected dial count 1, got %d", dialCount)
		}
	})

	t.Run("prevents multiple dials", func(t *testing.T) {
		dialCount := 0
		mockDial := func(ctx context.Context, network, address string) (net.Conn, error) {
			dialCount++
			return &mockConn{}, nil
		}

		wrapped := Wrap(mockDial)

		// First dial should succeed
		conn1, err1 := wrapped(context.Background(), "tcp", "example.com:80")
		if err1 != nil {
			t.Errorf("first dial: unexpected error: %v", err1)
		}
		if conn1 == nil {
			t.Error("first dial: expected connection, got nil")
		}

		// Second dial should fail
		conn2, err2 := wrapped(context.Background(), "tcp", "example.com:80")
		if !errors.Is(err2, ErrMultipleDial) {
			t.Errorf("second dial: expected ErrMultipleDial, got %v", err2)
		}
		if conn2 != nil {
			t.Error("second dial: expected nil connection")
		}
		if dialCount != 1 {
			t.Errorf("expected dial count 1, got %d", dialCount)
		}
	})

	t.Run("preserves underlying dial errors", func(t *testing.T) {
		expectedErr := errors.New("dial error")
		mockDial := func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, expectedErr
		}

		wrapped := Wrap(mockDial)
		conn, err := wrapped(context.Background(), "tcp", "example.com:80")

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if conn != nil {
			t.Error("expected nil connection")
		}
	})
}
