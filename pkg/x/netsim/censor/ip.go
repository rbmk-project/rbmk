// SPDX-License-Identifier: GPL-3.0-or-later

package censor

import (
	"bytes"
	"net/netip"
	"sync"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// Blackholer implements connection blackholing with optional pattern matching
// and connection tracking. Once a connection is blackholed, all packets matching
// its five-tuple will be dropped for the configured duration.
type Blackholer struct {
	// target specifies an optional specific endpoint to filter
	// if zero, applies to all connections.
	target netip.AddrPort

	// pattern is an optional byte pattern to match in payload
	// if nil, only considers the target (if set).
	pattern []byte

	// duration specifies how long to maintain blackholing state, if set.
	duration time.Duration

	// mu protects access to blocked.
	mu sync.Mutex

	// blocked tracks blackholed connections using five-tuple.
	blocked map[fiveTuple]time.Time
}

// fiveTuple is the five-tuple identifying a connection.
type fiveTuple struct {
	proto   packet.IPProtocol
	srcAddr netip.Addr
	srcPort uint16
	dstAddr netip.Addr
	dstPort uint16
}

// NewBlackholer creates a new [*Blackholer] instance.
//
// The duration parameter controls how long connections remain blackholed.
//
// If target is zero, it applies to all connections.
//
// If pattern is nil, it doesn't perform payload matching.
func NewBlackholer(duration time.Duration, target netip.AddrPort, pattern []byte) *Blackholer {
	return &Blackholer{
		target:   target,
		pattern:  pattern,
		duration: duration,
		mu:       sync.Mutex{},
		blocked:  make(map[fiveTuple]time.Time),
	}
}

// Filter implements [packet.Filter].
func (t *Blackholer) Filter(pkt *packet.Packet) (packet.Target, []*packet.Packet) {
	// Check if this connection is already blocked
	tuple := fiveTuple{
		proto:   pkt.IPProtocol,
		srcAddr: pkt.SrcAddr,
		srcPort: pkt.SrcPort,
		dstAddr: pkt.DstAddr,
		dstPort: pkt.DstPort,
	}
	now := time.Now()
	t.mu.Lock()
	deadline, ok := t.blocked[tuple]
	blocked := ok && now.Before(deadline)
	if ok && !blocked {
		delete(t.blocked, tuple)
	}
	t.mu.Unlock()
	if blocked {
		return packet.DROP, nil
	}

	// Check if we need to filter specific endpoint
	if t.target.IsValid() {
		if pkt.DstAddr != t.target.Addr() || pkt.DstPort != t.target.Port() {
			return packet.CONTINUE, nil
		}
	}

	// If we have a pattern, check payload
	if t.pattern != nil {
		if len(pkt.Payload) <= 0 || !bytes.Contains(pkt.Payload, t.pattern) {
			return packet.CONTINUE, nil
		}
	}

	// Block this connection
	t.mu.Lock()
	t.blocked[tuple] = now.Add(t.duration)
	t.mu.Unlock()

	return packet.DROP, nil
}

// DNatter implements transparent proxying via DNAT (Destination NAT).
type DNatter struct {
	// source is the source address to DNAT.
	source netip.Addr

	// target is the target destination endpoint to replace.
	target netip.AddrPort

	// repl is the replacement destination endpoint.
	repl netip.AddrPort
}

// NewDNatter creates a new [*DNatter] instance.
//
// Arguments:
//
// - source is the source address to DNAT.
//
// - target is the target destination endpoint to replace.
//
// - repl is the replacement destination endpoint.
//
// For example, with:
//
// - source = "193.206.158.22"
//
// - target = "93.184.216.34:80"
//
// - repl = "10.10.34.35:80"
//
// Traffic from "192.206.168.22" to "93.184.216.34:80" will be sent
// to "10.10.34.35:80" instead and return traffic from "10.10.345.35:80" to
// "192.206.168.22" would seem to come from "93.184.216.34:80".
func NewDNatter(source netip.Addr, target, repl netip.AddrPort) *DNatter {
	return &DNatter{
		source: source,
		target: target,
		repl:   repl,
	}
}

// Filter implements [packet.Filter].
func (r *DNatter) Filter(pkt *packet.Packet) (packet.Target, []*packet.Packet) {
	// forward match on the DNAT rule
	if pkt.SrcAddr == r.source && (pkt.DstAddr == r.target.Addr() && pkt.DstPort == r.target.Port()) {
		pkt.DstAddr = r.repl.Addr()
		pkt.DstPort = r.repl.Port()
		return packet.CONTINUE, nil
	}

	// return patch match on the DNAT rule
	if (pkt.SrcAddr == r.repl.Addr() && pkt.SrcPort == r.repl.Port()) && pkt.DstAddr == r.source {
		pkt.SrcAddr = r.target.Addr()
		pkt.SrcPort = r.target.Port()
		return packet.CONTINUE, nil
	}

	// otherwise just accept the packet
	return packet.CONTINUE, nil
}
