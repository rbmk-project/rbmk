//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TCP Listener/PacketListener.
//

package netstack

import (
	"net"
	"net/netip"
)

// TCPListenerStack is the stack to which a [*TCPListener] is attached.
type TCPListenerStack interface {
	NewTCPConn(laddr, raddr netip.AddrPort) (*TCPConn, error)
}

// TCPListener is a TCP listener.
//
// The zero value is invalid; construct using [NewTCPListener].
type TCPListener struct {
	*Port
	stack TCPListenerStack
}

// NewTCPListener creates a new TCP connection.
func NewTCPListener(stack TCPListenerStack, p *Port) *TCPListener {
	return &TCPListener{stack: stack, Port: p}
}

// Ensure [*TCPListener] implements [net.Listener].
var _ net.Listener = &TCPListener{}

// Accept implements [net.Listener].
func (tl *TCPListener) Accept() (net.Conn, error) {
	for {
		// Await for incoming packets or stop when done
		pkt, err := tl.Port.ReadPacket()
		if err != nil {
			return nil, err
		}

		// Expect SYN and respond with SYN|ACK flags
		if pkt.Flags != TCPFlagSYN {
			continue
		}
		laddr := netip.AddrPortFrom(pkt.DstAddr, pkt.DstPort)
		raddr := netip.AddrPortFrom(pkt.SrcAddr, pkt.SrcPort)
		conn, err := tl.stack.NewTCPConn(laddr, raddr)
		if err != nil {
			continue
		}
		if err := conn.Accept(); err != nil {
			conn.Close()
			continue
		}
		return conn, nil
	}
}

// Addr implements [net.Listener].
func (tl *TCPListener) Addr() net.Addr {
	return tl.Port.LocalAddr()
}
