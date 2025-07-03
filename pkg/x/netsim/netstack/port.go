//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TCP/UDP port implementation.
//

package netstack

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"
)

// PortAddr is the [*Port] address.
type PortAddr struct {
	// LocalAddr is the local address. This field must
	// always have valid address and port.
	LocalAddr netip.AddrPort

	// Protocol is the port protocol.
	Protocol IPProtocol

	// RemoteAddr is the remote address. This field
	// may be zero for non-connected ports.
	RemoteAddr netip.AddrPort
}

// String returns the string representation of the [*PortAddr].
func (pa *PortAddr) String() string {
	raddr := pa.RemoteAddr.String()
	if !pa.RemoteAddr.IsValid() {
		raddr = "*:*"
	}
	return fmt.Sprintf("%s -> %s %s", pa.LocalAddr, raddr, pa.Protocol)
}

// PortStack is the stack to which a [*Port] is attached.
type PortStack interface {
	// ClosePort closes the given port.
	ClosePort(addr *PortAddr)

	// FindLocalAddrFor finds the first local address with the
	// same family of the given remote address.
	FindLocalAddrFor(raddr netip.Addr) (netip.Addr, error)
}

// Port models an open TCP/UDP port.
type Port struct {
	// addr contains the port address.
	addr *PortAddr

	// eof unblocks any pending read.
	eof chan struct{}

	// eofOnce ensures we close just once.
	eofOnce sync.Once

	// input is the channel where we receive input.
	input chan *Packet

	// output is the channel where we post output.
	output chan *Packet

	// rd is the deadline for read operations.
	rd *deadline

	// stack is the underlying net stack.
	stack PortStack

	// wd is the deadline for write operations.
	wd *deadline
}

// NewPort creates a [*Port] instance with the given [*PortAddr].
//
// Leave the [*PortAddr] `RemoteAddr` field zero when you want to create
// a port that is not connected to a peer (i.e., a TCP/UDP listener).
func NewPort(stack PortStack, addr *PortAddr) *Port {
	return &Port{
		addr:    addr,
		eof:     make(chan struct{}),
		eofOnce: sync.Once{},
		input:   make(chan *Packet),
		output:  make(chan *Packet),
		rd:      newDeadline(),
		stack:   stack,
		wd:      newDeadline(),
	}
}

// Close closes the [*Port] terminating any pending I/O.
func (up *Port) Close() error {
	up.eofOnce.Do(func() {
		up.stack.ClosePort(up.addr)
		close(up.eof)
		up.rd.Set(time.Time{})
		up.wd.Set(time.Time{})
	})
	return nil
}

// LocalAddr returns the local address of this [*Port].
func (gp *Port) LocalAddr() net.Addr {
	return &Addr{gp.addr.LocalAddr, gp.addr.Protocol}
}

// RemoteAddr returns the remote address of this [*Port].
func (gp *Port) RemoteAddr() net.Addr {
	return &Addr{gp.addr.RemoteAddr, gp.addr.Protocol}
}

// SetDeadline sets the read and write deadlines.
func (gp *Port) SetDeadline(t time.Time) error {
	gp.SetReadDeadline(t)
	gp.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline sets the read deadline.
func (gp *Port) SetReadDeadline(t time.Time) error {
	gp.rd.Set(t)
	return nil
}

// SetWriteDeadline sets the write deadline.
func (gp *Port) SetWriteDeadline(t time.Time) error {
	gp.wd.Set(t)
	return nil
}

// ReadFrom implements [net.PacketConn].
func (gp *Port) ReadFrom(buf []byte) (int, net.Addr, error) {
	pkt, err := gp.ReadPacket()
	if err != nil {
		return 0, nil, err
	}
	count := copy(buf, pkt.Payload)
	srcAddr := netip.AddrPortFrom(pkt.SrcAddr, pkt.SrcPort)
	return count, &Addr{srcAddr, pkt.IPProtocol}, nil
}

// ReadPacket receives a packet from a remote endpoint.
//
// We discard packets that do not match the remote address unless the
// remote address is not set, in which case we accept all packets.
//
// The following errors are possible:
//
// 1. nil if we receive a packet from the `Input` channel.
//
// 2. [net.ErrClosed] if the port is closed before we receive a packet;
//
// 3. [os.ErrDeadlineExceeded] if the read deadline is exceeded.
func (gp *Port) ReadPacket() (*Packet, error) {
	for {
		select {
		case pkt := <-gp.input:
			// As documented, discard non-matching packets
			if !gp.addr.RemoteAddr.IsValid() || pkt.SrcAddr == gp.addr.RemoteAddr.Addr() {
				return pkt, nil
			}

		case <-gp.eof:
			return nil, net.ErrClosed

		case <-gp.rd.Wait():
			return nil, os.ErrDeadlineExceeded
		}
	}
}

// WriteTo implements [net.PacketConn].
func (gp *Port) WriteTo(pkt []byte, addr net.Addr) (int, error) {
	raddr, err := netip.ParseAddrPort(addr.String())
	if err != nil {
		return 0, EINVAL
	}
	if err := gp.WritePacket(pkt, 0, raddr); err != nil {
		return 0, err
	}
	return len(pkt), nil
}

// Write implements [net.Conn].
func (gp *Port) Write(data []byte) (int, error) {
	if err := gp.WritePacket(data, 0, netip.AddrPort{}); err != nil {
		return 0, err
	}
	return len(data), nil
}

// WritePacket creates and writes a packet to a remote endpoint using
// the given payload, TCP flags, and remote address.
//
// If the `raddr` field is a zero value, we use the `RemoteAddr`
// field of the [*PortAddr]. If also such a field is a zero value,
// we return [ENOTCONN] to indicate we don't know the peer addr.
//
// Also, we copy the payload to avoid issues with buffer pools, which
// occur, for example, when using the [crypto/tls] package.
//
// The following errors are possible:
//
// 1. [ENOTCONN] if the port is not connected to a peer and the raddr is zero;
//
// 2. nil if the packet is sent (i.e., delivered to the `Output` channel);
//
// 3. [net.ErrClosed] if the port is closed before we send the packet;
//
// 4. [os.ErrDeadlineExceeded] if the write deadline is exceeded.
func (gp *Port) WritePacket(payload []byte, flags TCPFlags, raddr netip.AddrPort) error {
	// Attempt to figure out the remote address first
	if !raddr.IsValid() {
		raddr = gp.addr.RemoteAddr
		if !raddr.IsValid() {
			return ENOTCONN
		}
	}

	// Make sure the local address is specified
	laddr := gp.addr.LocalAddr
	if laddr.Addr().IsUnspecified() {
		ipAddr, err := gp.stack.FindLocalAddrFor(raddr.Addr())
		if err != nil {
			return err
		}
		laddr = netip.AddrPortFrom(ipAddr, laddr.Port())
	}

	// Build and send the packet.
	//
	// As documented, copy the payload.
	const linuxDefaultTTL = 64
	pkt := &Packet{
		TTL:        linuxDefaultTTL,
		SrcAddr:    laddr.Addr(),
		DstAddr:    raddr.Addr(),
		IPProtocol: gp.addr.Protocol,
		SrcPort:    laddr.Port(),
		DstPort:    raddr.Port(),
		Flags:      flags,
		Payload:    append([]byte{}, payload...),
	}
	select {
	case gp.output <- pkt:
		return nil
	case <-gp.eof:
		return net.ErrClosed
	case <-gp.wd.Wait():
		return os.ErrDeadlineExceeded
	}
}

// Input returns the channel to write to send a [*Packet] to the [*Port].
func (gp *Port) Input() chan<- *Packet {
	return gp.input
}

// Output returns the channel to read to recv a [*Packet] from the [*Port].
func (gp *Port) Output() <-chan *Packet {
	return gp.output
}
