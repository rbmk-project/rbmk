// SPDX-License-Identifier: GPL-3.0-or-later

package mocks

import (
	"net"
	"time"
)

// PacketConn is a mockable [net.PacketConn].
type PacketConn struct {
	// MockReadFrom is the function to call when ReadFrom is called.
	MockReadFrom func(p []byte) (int, net.Addr, error)

	// MockWriteTo is the function to call when WriteTo is called.
	MockWriteTo func(p []byte, addr net.Addr) (int, error)

	// MockClose is the function to call when Close is called.
	MockClose func() error

	// MockLocalAddr is the function to call when LocalAddr is called.
	MockLocalAddr func() net.Addr

	// MockSetDeadline is the function to call when SetDeadline is called.
	MockSetDeadline func(t time.Time) error

	// MockSetReadDeadline is the function to call when SetReadDeadline is called.
	MockSetReadDeadline func(t time.Time) error

	// MockSetWriteDeadline is the function to call when SetWriteDeadline is called.
	MockSetWriteDeadline func(t time.Time) error
}

var _ net.PacketConn = &PacketConn{}

// ReadFrom calls MockReadFrom.
func (pc *PacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	return pc.MockReadFrom(p)
}

// WriteTo calls MockWriteTo.
func (pc *PacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	return pc.MockWriteTo(p, addr)
}

// Close calls MockClose.
func (pc *PacketConn) Close() error {
	return pc.MockClose()
}

// LocalAddr calls MockLocalAddr.
func (pc *PacketConn) LocalAddr() net.Addr {
	return pc.MockLocalAddr()
}

// SetDeadline calls MockSetDeadline.
func (pc *PacketConn) SetDeadline(t time.Time) error {
	return pc.MockSetDeadline(t)
}

// SetReadDeadline calls MockSetReadDeadline.
func (pc *PacketConn) SetReadDeadline(t time.Time) error {
	return pc.MockSetReadDeadline(t)
}

// SetWriteDeadline calls MockSetWriteDeadline.
func (pc *PacketConn) SetWriteDeadline(t time.Time) error {
	return pc.MockSetWriteDeadline(t)
}
