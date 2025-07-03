// SPDX-License-Identifier: GPL-3.0-or-later

package censor

import (
	"bytes"
	"net/netip"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// TCPResetter implements RST-based TCP connection interruption.
//
// When configured with a pattern, it only injects RST segments
// for packets containing that pattern, while allowing empty
// packets (e.g., SYN) to pass through. This enables pattern matching
// on protocol-specific content (e.g., TLS SNI) while allowing
// the TCP handshake to complete normally.
type TCPResetter struct {
	// target specifies an optional specific endpoint to filter;
	// if zero, applies to all TCP connections.
	target netip.AddrPort

	// pattern is an optional byte pattern to match in payload;
	// if nil, only considers the target (if set).
	pattern []byte
}

// NewTCPResetter creates a new [*TCPResetter].
//
// If target is zero, it applies to all TCP connections.
//
// If pattern is zero-length, it doesn't perform payload matching.
//
// When pattern is set, empty packets are allowed through
// to permit TCP handshakes to complete.
func NewTCPResetter(target netip.AddrPort, pattern []byte) *TCPResetter {
	return &TCPResetter{target: target, pattern: pattern}
}

// Filter implements [packet.Filter].
func (r *TCPResetter) Filter(pkt *packet.Packet) (packet.Target, []*packet.Packet) {
	// Only process TCP packets
	if pkt.IPProtocol != packet.IPProtocolTCP {
		return packet.CONTINUE, nil
	}

	// Check if we need to filter a specific endpoint
	if r.target.IsValid() {
		if pkt.DstAddr != r.target.Addr() || pkt.DstPort != r.target.Port() {
			return packet.CONTINUE, nil
		}
	}

	// If we have a pattern, check the payload. Note: we explicitly
	// accept packets with empty payload (e.g., SYN) to allow the TCP
	// handshake to complete before potentially injecting RST.
	if r.pattern != nil {
		if len(pkt.Payload) <= 0 || !bytes.Contains(pkt.Payload, r.pattern) {
			return packet.CONTINUE, nil
		}
	}

	// Create RST packet
	rst := &packet.Packet{
		TTL:        64,
		SrcAddr:    pkt.DstAddr,
		DstAddr:    pkt.SrcAddr,
		IPProtocol: packet.IPProtocolTCP,
		SrcPort:    pkt.DstPort,
		DstPort:    pkt.SrcPort,
		Flags:      packet.TCPFlagRST,
	}

	return packet.CONTINUE, []*packet.Packet{rst}
}
