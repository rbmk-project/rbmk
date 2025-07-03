// SPDX-License-Identifier: GPL-3.0-or-later

// Package link models a point-to-point network link.
package link

import (
	"sync"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// LinkStack is the [*Stack] as seen by a [*Link].
type LinkStack = packet.NetworkDevice

// Packet is the [packet.Packet] alias used by this package.
type Packet = packet.Packet

// Link models a link between two [*Stack] instances.
//
// The zero value is not ready to use; construct using [New].
type Link struct {
	// eof unblocks any blocking channel operation.
	eof chan struct{}

	// eofOnce ensures we close just once.
	eofOnce sync.Once
}

// New creates a new [*Link] using two [*Stack] and
// sets up moving packets between the two stacks. Use Close
// to shut down background goroutines.
func New(left, right LinkStack) *Link {
	lnk := &Link{
		eof:     make(chan struct{}),
		eofOnce: sync.Once{},
	}
	go lnk.move(left, right)
	go lnk.move(right, left)
	return lnk
}

// Close stops background goroutines moving traffic.
func (lnk *Link) Close() error {
	lnk.eofOnce.Do(func() { close(lnk.eof) })
	return nil
}

type readableStack interface {
	EOF() <-chan struct{}
	Output() <-chan *Packet
}

type writableStack interface {
	EOF() <-chan struct{}
	Input() chan<- *Packet
}

// move moves packets from the left stack to the right stack.
func (lnk *Link) move(left readableStack, right writableStack) {
	for {
		// Read from left stack.
		select {
		case <-lnk.eof:
			return
		case <-left.EOF():
			return
		case pkt := <-left.Output():

			// Write to right stack.
			select {
			case <-lnk.eof:
				return
			case <-right.EOF():
				return
			case right.Input() <- pkt:
				// success
			}
		}
	}
}
