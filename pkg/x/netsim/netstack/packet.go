//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Aliases for Packet and the related definitions.
//

package netstack

import "github.com/rbmk-project/rbmk/pkg/x/netsim/packet"

// Type aliases
type (
	Packet     = packet.Packet
	IPProtocol = packet.IPProtocol
	TCPFlags   = packet.TCPFlags
)

// Constant aliases
const (
	IPProtocolTCP = packet.IPProtocolTCP
	IPProtocolUDP = packet.IPProtocolUDP

	TCPFlagFIN = packet.TCPFlagFIN
	TCPFlagSYN = packet.TCPFlagSYN
	TCPFlagRST = packet.TCPFlagRST
	TCPFlagPSH = packet.TCPFlagPSH
	TCPFlagACK = packet.TCPFlagACK
)
