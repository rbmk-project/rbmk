// SPDX-License-Identifier: GPL-3.0-or-later

// Package netipx contains [net/netip] extensions.
package netipx

import (
	"net"
	"net/netip"
)

// AddrToAddrPort converts a [net.Addr] to a [netip.AddrPort].
//
// If the input is nil or neither a [*net.TCPAddr] nor [*net.UDPAddr],
// returns an unspecified IPv6 address with port 0.
//
// For [*net.TCPAddr] and [*net.UDPAddr] addresses, returns their
// corresponding [netip.AddrPort] representation.
func AddrToAddrPort(addr net.Addr) netip.AddrPort {
	if addr == nil {
		return netip.AddrPortFrom(netip.IPv6Unspecified(), 0)
	}
	if tcp, ok := addr.(*net.TCPAddr); ok {
		return tcp.AddrPort()
	}
	if udp, ok := addr.(*net.UDPAddr); ok {
		return udp.AddrPort()
	}
	return netip.AddrPortFrom(netip.IPv6Unspecified(), 0)
}
