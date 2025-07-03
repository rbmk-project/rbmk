//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Network stack
//

package netstack

import (
	"context"
	"log"
	"math"
	"net"
	"net/netip"
	"sync"

	"github.com/rbmk-project/rbmk/pkg/common/runtimex"
	"github.com/rbmk-project/rbmk/pkg/dns/dnscore"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// Stack models a network stack.
type Stack struct {
	// addr is the stack network address.
	addrs []netip.Addr

	// eof unblocks any blocking operation when the stack is closed.
	eof chan struct{}

	// eofOnce ensures we close just once.
	eofOnce sync.Once

	// input is the input channel for packets.
	input chan *Packet

	// nextport tracks the next available ephemeral port.
	nextport map[IPProtocol]uint16

	// output is the output channel for packets.
	output chan *Packet

	// portmu protects nextport and ports.
	portmu sync.RWMutex

	// resolvers contains the DNS resolvers to use.
	resolvers []*dnscore.ServerAddr

	// ports contains the open ports.
	ports map[PortAddr]*Port
}

// New creates a new [*Stack] instance and starts a
// goroutine demuxing incoming traffic. Remember to invoke
// Close to stop any muxing/demuxing goroutine.
func New(addrs ...netip.Addr) *Stack {
	const (
		// firstEphemeralPort is the first ephemeral port
		// to use according to RFC6335.
		firstEphemeralPort = 49152
	)
	// We use buffered channels for I/O because that allows
	// routers to use nonblocking writes w/o dropping packets
	// unless the receiver is actively ignoring them.
	//
	// The buffer also allows us to send RST after SYN to
	// closed port using nonblocking I/O.
	input, output := packet.NewNetworkDeviceIOChannels()
	ns := &Stack{
		addrs:   addrs,
		eof:     make(chan struct{}),
		eofOnce: sync.Once{},
		input:   input,
		nextport: map[IPProtocol]uint16{
			IPProtocolTCP: firstEphemeralPort,
			IPProtocolUDP: firstEphemeralPort,
		},
		output: output,
		portmu: sync.RWMutex{},
		ports:  map[PortAddr]*Port{},
	}
	go ns.demuxLoop()
	return ns
}

// SetResolvers sets the resolvers endpoints to use.
//
// Note that this method IS NOT goroutine safe.
func (ns *Stack) SetResolvers(addrs ...netip.AddrPort) {
	ns.resolvers = nil
	for _, addr := range addrs {
		ns.resolvers = append(ns.resolvers, &dnscore.ServerAddr{
			Protocol: dnscore.ProtocolUDP,
			Address:  addr.String(),
		})
	}
}

// Addresses returns the network stack addresses.
func (ns *Stack) Addresses() []netip.Addr {
	return append([]netip.Addr{}, ns.addrs...)
}

// EOF returns the channel to wait for the stack to close.
func (ns *Stack) EOF() <-chan struct{} {
	return ns.eof
}

// demuxLoop demuxes incoming traffic to the proper port.
func (ns *Stack) demuxLoop() {
	for {
		select {
		case <-ns.eof:
			return
		case pkt := <-ns.input:
			ns.demux(pkt)
		}
	}
}

// findPortLocked finds a port using the given address.
//
// The algorithm is as follows:
//
// 1. first, try using the five tuple.
//
// 2. if not found, try using the three tuple, where
// the remote address is invalid.
//
// 3. if not found, use a five tuple where the
// local IP address is unspecified.
//
// 4. if not found, use a three tuple where the
// the remote address is invalid, and the IP local
// address is unspecified.
//
// 5. otherwise, return nil.
//
// The caller must hold the portmu lock.
func (ns *Stack) findPortLocked(pkt *Packet) *Port {
	// 1.
	addr := PortAddr{
		LocalAddr:  netip.AddrPortFrom(pkt.DstAddr, pkt.DstPort),
		Protocol:   pkt.IPProtocol,
		RemoteAddr: netip.AddrPortFrom(pkt.SrcAddr, pkt.SrcPort),
	}
	if port := ns.ports[addr]; port != nil {
		return port
	}

	// 2.
	addr = PortAddr{
		LocalAddr:  netip.AddrPortFrom(pkt.DstAddr, pkt.DstPort),
		Protocol:   pkt.IPProtocol,
		RemoteAddr: netip.AddrPort{},
	}
	if port := ns.ports[addr]; port != nil {
		return port
	}

	for _, ipAddr := range []netip.Addr{netip.IPv4Unspecified(), netip.IPv6Unspecified()} {
		// 3.
		addr = PortAddr{
			LocalAddr:  netip.AddrPortFrom(ipAddr, pkt.DstPort),
			Protocol:   pkt.IPProtocol,
			RemoteAddr: netip.AddrPortFrom(pkt.SrcAddr, pkt.SrcPort),
		}
		if port := ns.ports[addr]; port != nil {
			return port
		}

		// 4.
		addr = PortAddr{
			LocalAddr:  netip.AddrPortFrom(ipAddr, pkt.DstPort),
			Protocol:   pkt.IPProtocol,
			RemoteAddr: netip.AddrPort{},
		}
		if port := ns.ports[addr]; port != nil {
			return port
		}
	}

	return nil
}

// resetNonblocking sends a RST packet in response to a SYN for a closed port.
func (ns *Stack) resetNonblocking(pkt *Packet) {
	runtimex.Assert(pkt.IPProtocol == IPProtocolTCP, "not a TCP packet")
	runtimex.Assert(pkt.Flags == TCPFlagSYN, "expected SYN flags")
	resp := &Packet{
		SrcAddr:    pkt.DstAddr,
		DstAddr:    pkt.SrcAddr,
		IPProtocol: IPProtocolTCP,
		SrcPort:    pkt.DstPort,
		DstPort:    pkt.SrcPort,
		Flags:      TCPFlagRST,
		Payload:    []byte{},
	}
	// Nonblocking write to the buffered output channel.
	select {
	case ns.output <- resp:
	default:
	}
}

// demux demuxes a single incoming [*Packet].
func (ns *Stack) demux(pkt *Packet) error {
	// Discard packet if the TTL is zero.
	if pkt.TTL <= 0 {
		return EHOSTUNREACH
	}

	// Discard packet if the address is not local.
	if !ns.isLocalAddr(pkt.DstAddr) {
		return EHOSTUNREACH
	}

	// Find a route using the five tuple then fallback using
	// the three tuple for listening sockets.
	ns.portmu.RLock()
	port := ns.findPortLocked(pkt)
	ns.portmu.RUnlock()
	if port == nil {
		if pkt.IPProtocol == IPProtocolTCP && pkt.Flags == TCPFlagSYN {
			ns.resetNonblocking(pkt)
		}
		return ECONNREFUSED
	}

	// Actually deliver the packet to the port.
	select {
	case <-port.eof:
		return net.ErrClosed
	case <-ns.eof:
		return ENETDOWN
	case port.input <- pkt:
		return nil
	}
}

// Close closes the network stack and stops all traffic muxing/demuxing.
func (ns *Stack) Close() error {
	ns.eofOnce.Do(func() { close(ns.eof) })
	return nil
}

// Output returns the channel from which to read outgoing packets.
func (ns *Stack) Output() <-chan *Packet {
	return ns.output
}

// Input returns the channel where to write incoming packets.
func (ns *Stack) Input() chan<- *Packet {
	return ns.input
}

// ListenPacket creates a new listening [net.PacketConn].
func (ns *Stack) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	if network != "udp" {
		return nil, EPROTONOSUPPORT
	}
	port, err := ns.listen(IPProtocolUDP, address)
	if err != nil {
		return nil, err
	}
	return NewUDPConn(port), nil
}

// Listen creates a new listening [net.Listener].
func (ns *Stack) Listen(ctx context.Context, network, address string) (net.Listener, error) {
	if network != "tcp" {
		return nil, EPROTONOSUPPORT
	}
	port, err := ns.listen(IPProtocolTCP, address)
	if err != nil {
		return nil, err
	}
	return NewTCPListener(ns, port), nil
}

// dialContext is the internal function to dial that only accepts
// addresses containing IPv4 or IPv6 addresses and a port.
func (ns *Stack) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	case "udp":
		port, err := ns.dial(IPProtocolUDP, address)
		if err != nil {
			return nil, err
		}
		return NewUDPConn(port), nil

	case "tcp":
		port, err := ns.dial(IPProtocolTCP, address)
		if err != nil {
			return nil, err
		}
		conn := NewTCPConn(port)
		if err := conn.Connect(ctx); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil

	default:
		return nil, EPROTONOSUPPORT
	}
}

// isLocalAddr returns true if the address is local to the stack.
func (ns *Stack) isLocalAddr(addr netip.Addr) bool {
	for _, a := range ns.addrs {
		if a == addr {
			return true
		}
	}
	return false
}

// listen creates a new listening [*Port].
func (ns *Stack) listen(protocol IPProtocol, address string) (*Port, error) {
	// Run while locking the available ports.
	ns.portmu.Lock()
	defer ns.portmu.Unlock()

	// Setup the local address handling the cases in which the
	// address and/or the port are the zero value.
	laddr, err := netip.ParseAddrPort(address)
	if err != nil {
		return nil, EINVAL
	}
	if !laddr.Addr().IsUnspecified() && !ns.isLocalAddr(laddr.Addr()) {
		return nil, EADDRNOTAVAIL
	}
	if laddr.Port() <= 0 {
		lport, err := ns.newEphemeralPortNumberLocked(protocol)
		if err != nil {
			return nil, err
		}
		laddr = netip.AddrPortFrom(laddr.Addr(), lport)
	}

	// The remote address is always unspecified in this case.
	var raddr netip.AddrPort

	// Create the port proper and setup muxing traffic.
	return ns.newPortLocked(protocol, laddr, raddr)
}

// dial creates a new connected [*Port].
func (ns *Stack) dial(protocol IPProtocol, address string) (*Port, error) {
	// Run while locking the available ports.
	ns.portmu.Lock()
	defer ns.portmu.Unlock()

	// Setup the remote address and make sure it is actually specified.
	raddr, err := netip.ParseAddrPort(address)
	if err != nil {
		return nil, EINVAL
	}
	if raddr.Addr().IsUnspecified() || raddr.Port() <= 0 {
		return nil, EHOSTUNREACH
	}

	// Pick the correct local address for the remote address.
	ipAddrLocal, err := ns.FindLocalAddrFor(raddr.Addr())
	if err != nil {
		return nil, err
	}

	// Construct the local address and use a local port.
	lport, err := ns.newEphemeralPortNumberLocked(protocol)
	if err != nil {
		return nil, err
	}
	laddr := netip.AddrPortFrom(ipAddrLocal, lport)

	// Create the port proper and setup muxing traffic.
	return ns.newPortLocked(protocol, laddr, raddr)
}

// FindLocalAddrFor returns the first local address that has
// the same IP version as the remote address.
func (ns *Stack) FindLocalAddrFor(raddr netip.Addr) (netip.Addr, error) {
	// Normalize to IPv4 if mapped.
	if raddr.Is4In6() {
		raddr = raddr.Unmap()
	}

	var result netip.Addr
	for _, addr := range ns.addrs {
		// Normalize to IPv4 if mapped.
		if addr.Is4In6() {
			addr = addr.Unmap()
		}

		// Check for exact match with address family
		if raddr.Is4() && addr.Is4() {
			result = addr
			break
		}
		if raddr.Is6() && addr.Is6() {
			result = addr
			break
		}
	}

	// Produce the final result.
	if !result.IsValid() {
		return netip.Addr{}, EADDRNOTAVAIL
	}
	return result, nil
}

// newEphemeralPortNumberLocked opens a new local port, if possible, or returns an error.
//
// You must invoke this method while holding the portmu lock.
func (ns *Stack) newEphemeralPortNumberLocked(protocol IPProtocol) (uint16, error) {
	if ns.nextport[protocol] >= math.MaxUint16 {
		return 0, EADDRINUSE
	}
	port := ns.nextport[protocol]
	ns.nextport[protocol] = port + 1
	return port, nil
}

// newPortLocked creates a new [*Port] instance.
//
// You must invoke this method while holding the portmu lock.
func (ns *Stack) newPortLocked(protocol IPProtocol, laddr, raddr netip.AddrPort) (*Port, error) {
	// Create the port address and make sure we can actually create the port.
	addr := &PortAddr{
		LocalAddr:  laddr,
		Protocol:   protocol,
		RemoteAddr: raddr,
	}
	port := NewPort(ns, addr)
	if _, ok := ns.ports[*addr]; ok {
		return nil, EADDRINUSE
	}

	// Remember the port and routing traffic
	log.Printf("OPEN %s", addr.String())
	ns.ports[*addr] = port
	go ns.muxOutgoingTraffic(port)
	return port, nil
}

// muxOutgoingTraffic merges the traffic emitted by all ports.
func (ns *Stack) muxOutgoingTraffic(port *Port) {
	for {
		select {
		case <-port.eof:
			return
		case <-ns.eof:
			return
		case pkt := <-port.output:
			ns.output <- pkt
		}
	}
}

// ClosePort implements [PortStack].
func (ns *Stack) ClosePort(addr *PortAddr) {
	log.Printf("CLOSE %s", addr.String())
	ns.portmu.Lock()
	delete(ns.ports, *addr)
	ns.portmu.Unlock()
}

// NewTCPConn implements [TCPListenerStack].
func (ns *Stack) NewTCPConn(laddr, raddr netip.AddrPort) (*TCPConn, error) {
	// Run while locking the available ports.
	ns.portmu.Lock()
	defer ns.portmu.Unlock()

	// Create the port proper and setup muxing traffic.
	port, err := ns.newPortLocked(IPProtocolTCP, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return NewTCPConn(port), nil
}
