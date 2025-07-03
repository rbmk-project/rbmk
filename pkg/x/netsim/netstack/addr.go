//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// net.Addr implementation.
//

package netstack

import (
	"net"
	"net/netip"
)

// Addr represents a TCP/UDP address.
type Addr struct {
	// AddrPort is the endpoint address and port.
	AddrPort netip.AddrPort

	// Protocol is the endpoint protocol.
	Protocol IPProtocol
}

// Ensure [*Addr] implements [net.Addr].
var _ net.Addr = &Addr{}

// Network implements [net.Addr].
func (sa *Addr) Network() string {
	return sa.Protocol.String()
}

// String implements [net.Addr].
func (sa *Addr) String() string {
	return sa.AddrPort.String()
}
