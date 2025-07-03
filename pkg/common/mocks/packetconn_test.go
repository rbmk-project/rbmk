// SPDX-License-Identifier: GPL-3.0-or-later

package mocks

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestPacketConn(t *testing.T) {
	t.Run("ReadFrom", func(t *testing.T) {
		expected := errors.New("mocked error")
		expectedAddr := &net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 8080,
		}
		pc := &PacketConn{
			MockReadFrom: func(p []byte) (int, net.Addr, error) {
				return 0, expectedAddr, expected
			},
		}
		count, addr, err := pc.ReadFrom(make([]byte, 128))
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
		if count != 0 {
			t.Fatal("expected 0 bytes")
		}
		if diff := cmp.Diff(expectedAddr, addr); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("WriteTo", func(t *testing.T) {
		expected := errors.New("mocked error")
		addr := &net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 8080,
		}
		pc := &PacketConn{
			MockWriteTo: func(p []byte, addr net.Addr) (int, error) {
				return 0, expected
			},
		}
		count, err := pc.WriteTo(make([]byte, 128), addr)
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
		if count != 0 {
			t.Fatal("expected 0 bytes")
		}
	})

	t.Run("Close", func(t *testing.T) {
		expected := errors.New("mocked error")
		pc := &PacketConn{
			MockClose: func() error {
				return expected
			},
		}
		err := pc.Close()
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
	})

	t.Run("LocalAddr", func(t *testing.T) {
		expected := &net.UDPAddr{
			IP:   net.IPv6loopback,
			Port: 1234,
		}
		pc := &PacketConn{
			MockLocalAddr: func() net.Addr {
				return expected
			},
		}
		out := pc.LocalAddr()
		if diff := cmp.Diff(expected, out); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("SetDeadline", func(t *testing.T) {
		expected := errors.New("mocked error")
		pc := &PacketConn{
			MockSetDeadline: func(t time.Time) error {
				return expected
			},
		}
		err := pc.SetDeadline(time.Time{})
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
	})

	t.Run("SetReadDeadline", func(t *testing.T) {
		expected := errors.New("mocked error")
		pc := &PacketConn{
			MockSetReadDeadline: func(t time.Time) error {
				return expected
			},
		}
		err := pc.SetReadDeadline(time.Time{})
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
	})

	t.Run("SetWriteDeadline", func(t *testing.T) {
		expected := errors.New("mocked error")
		pc := &PacketConn{
			MockSetWriteDeadline: func(t time.Time) error {
				return expected
			},
		}
		err := pc.SetWriteDeadline(time.Time{})
		if !errors.Is(err, expected) {
			t.Fatal("not the error we expected")
		}
	})
}
