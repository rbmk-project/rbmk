// SPDX-License-Identifier: GPL-3.0-or-later

// Package router provides network routing capabilities for testing.
package router

import (
	"errors"
	"net/netip"
	"sync"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// Router provides routing capabilities.
type Router struct {
	// filtermu protects access to filters.
	filtermu sync.RWMutex

	// filters contains pre-routing packet filters.
	filters []packet.Filter

	// srt is the static routing table.
	srt map[netip.Addr]packet.NetworkDevice
}

// New creates a new [*Router].
func New() *Router {
	return &Router{
		filtermu: sync.RWMutex{},
		filters:  make([]packet.Filter, 0),
		srt:      make(map[netip.Addr]packet.NetworkDevice),
	}
}

// AddFilter adds a packet filter to the router.
func (r *Router) AddFilter(pf packet.Filter) {
	r.filtermu.Lock()
	r.filters = append(r.filters, pf)
	r.filtermu.Unlock()
}

// Attach attaches a [packet.NetworkDevice] to the [*Router] reading
// packets from the router and setting up routes for all the device
// addresses to correctly forward packets back to the device.
func (r *Router) Attach(dev packet.NetworkDevice) {
	for _, addr := range dev.Addresses() {
		r.srt[addr] = dev
	}
	go r.readLoop(dev)
}

// readLoop reads packets from a [packet.NetworkDevice] until EOF.
func (r *Router) readLoop(dev packet.NetworkDevice) {
	for {
		select {
		case <-dev.EOF():
			return
		case pkt := <-dev.Output():
			r.handle(pkt)
		}
	}
}

// handle handles a packet by applying filters and routing it.
func (r *Router) handle(pkt *packet.Packet) error {
	// Get a consistent view of filters
	r.filtermu.RLock()
	filters := make([]packet.Filter, len(r.filters))
	copy(filters, r.filters)
	r.filtermu.RUnlock()

	// Apply filters
	for _, pf := range filters {
		target, inject := pf.Filter(pkt)

		// Handle any packets to inject
		for _, p := range inject {
			_ = r.route(p)
		}

		// Stop processing if packet should be dropped
		switch target {
		case packet.DROP:
			return nil
		default:
			// Continue processing
		}
	}

	// Route the original packet if it wasn't dropped
	return r.route(pkt)
}

var (
	// errTTLExceeded is returned when a packet's TTL is exceeded.
	errTTLExceeded = errors.New("TTL exceeded in transit")

	// errNoRouteToHost is returned when there is no route to the host.
	errNoRouteToHost = errors.New("no route to host")

	// errBufferFull is returned when the buffer is full.
	errBufferFull = errors.New("buffer full")
)

// route routes a given packet to its destination.
func (r *Router) route(pkt *packet.Packet) error {
	// Decrement TTL.
	if pkt.TTL <= 0 {
		return errTTLExceeded
	}
	pkt.TTL--

	// Find next hop.
	nextHop := r.srt[pkt.DstAddr]
	if nextHop == nil {
		return errNoRouteToHost
	}

	// Forward packet (non-blocking).
	select {
	case nextHop.Input() <- pkt:
		return nil
	default:
		return errBufferFull
	}
}
