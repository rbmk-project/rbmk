// SPDX-License-Identifier: GPL-3.0-or-later

package censor

import (
	"net/netip"

	"github.com/miekg/dns"
	netsimdns "github.com/rbmk-project/rbmk/pkg/x/netsim/dns"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// Database is an alias for [netsimdns.Database].
type Database = netsimdns.Database

// DNSPoisoner implements GFW-style DNS poisoning
type DNSPoisoner struct {
	addrs map[netip.Addr]struct{}
	db    *Database
}

// NewDNSPoisoner creates a new DNS poisoner that injects
// responses as configured in the given database.
func NewDNSPoisoner(db *Database, addrs ...netip.Addr) *DNSPoisoner {
	am := make(map[netip.Addr]struct{}, len(addrs))
	for _, addr := range addrs {
		am[addr] = struct{}{}
	}
	return &DNSPoisoner{addrs: am, db: db}
}

// Filter implements [packet.Filter].
func (p *DNSPoisoner) Filter(pkt *packet.Packet) (packet.Target, []*packet.Packet) {
	// Only process UDP DNS queries
	if pkt.IPProtocol != packet.IPProtocolUDP || pkt.DstPort != 53 {
		return packet.CONTINUE, nil
	}

	// Check whether we should only filter
	// specific network addresses
	if len(p.addrs) > 0 {
		if _, ok := p.addrs[pkt.DstAddr]; !ok {
			return packet.CONTINUE, nil
		}
	}

	// Parse DNS query
	query := new(dns.Msg)
	if err := query.Unpack(pkt.Payload); err != nil {
		return packet.CONTINUE, nil
	}

	// Only process queries
	if query.Response || len(query.Question) != 1 {
		return packet.CONTINUE, nil
	}

	// Create poisoned response
	spoofed := p.spoof(pkt, query)

	// Let original query continue
	return packet.CONTINUE, spoofed
}

func (p *DNSPoisoner) spoof(
	pkt *packet.Packet, query *dns.Msg) []*packet.Packet {
	// Prepare the response
	resp := &dns.Msg{}
	resp.SetReply(query)

	// Get records from database
	q0 := query.Question[0]
	rrs, found := p.db.Lookup(q0.Qtype, q0.Name)
	if !found {
		return []*packet.Packet{}
	}
	resp.Answer = rrs

	// Pack the response
	payload, err := resp.Pack()
	if err != nil {
		return []*packet.Packet{}
	}

	// Create the spoofed packet
	return []*packet.Packet{{
		TTL:        64,
		SrcAddr:    pkt.DstAddr,
		DstAddr:    pkt.SrcAddr,
		IPProtocol: packet.IPProtocolUDP,
		SrcPort:    pkt.DstPort,
		DstPort:    pkt.SrcPort,
		Payload:    payload,
	}}
}
