//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// UDP Conn/PacketConn.
//

package netstack

import (
	"net"
	"time"
)

// UDPConn is a UDP connection.
//
// The zero value is invalid; construct using [NewUDPConn].
type UDPConn struct {
	p *Port
}

// NewUDPConn creates a new UDP connection.
func NewUDPConn(p *Port) *UDPConn {
	return &UDPConn{p: p}
}

// Ensure [*UDPConn] implements [net.PacketConn].
var _ net.PacketConn = &UDPConn{}

// Close implements [net.PacketConn].
func (c *UDPConn) Close() error {
	return c.p.Close()
}

// LocalAddr implements [net.PacketConn].
func (c *UDPConn) LocalAddr() net.Addr {
	return c.p.LocalAddr()
}

// ReadFrom implements [net.PacketConn].
func (c *UDPConn) ReadFrom(buf []byte) (int, net.Addr, error) {
	return c.p.ReadFrom(buf)
}

// SetDeadline implements [net.PacketConn].
func (c *UDPConn) SetDeadline(t time.Time) error {
	return c.p.SetDeadline(t)
}

// SetReadDeadline implements [net.PacketConn].
func (c *UDPConn) SetReadDeadline(t time.Time) error {
	return c.p.SetReadDeadline(t)
}

// SetWriteDeadline implements net.PacketConn.
func (c *UDPConn) SetWriteDeadline(t time.Time) error {
	return c.p.SetWriteDeadline(t)
}

// WriteTo implements net.PacketConn.
func (c *UDPConn) WriteTo(pkt []byte, addr net.Addr) (int, error) {
	return c.p.WriteTo(pkt, addr)
}

// Ensure [*UDPConn] implements [net.Conn].
var _ net.Conn = &UDPConn{}

// Read implements [net.Conn].
func (c *UDPConn) Read(buf []byte) (int, error) {
	count, _, err := c.p.ReadFrom(buf)
	return count, err
}

// RemoteAddr implements [net.Conn].
func (c *UDPConn) RemoteAddr() net.Addr {
	return c.p.RemoteAddr()
}

// Write implements [net.Conn].
func (c *UDPConn) Write(data []byte) (int, error) {
	return c.p.Write(data)
}
