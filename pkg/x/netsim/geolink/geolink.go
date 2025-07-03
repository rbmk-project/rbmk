//
// SPDX-License-Identifier: BSD-3-Clause
//
// Adapted from: https://github.com/ooni/netem/blob/main/linkfwdfull.go
//

// Package geolink models a geographic point-to-point link.
package geolink

import (
	"log"
	"net/netip"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
)

// Config configures a geographic point-to-point link.
type Config struct {
	// Delay is the propagation delay.
	Delay time.Duration

	// Log enables logging of delivered packets.
	Log bool
}

// baseDevice is the common implementation for the
// devices type returned by this package.
type baseDevice struct {
	addresses []netip.Addr
	input     chan *packet.Packet
	output    chan *packet.Packet
}

func (dev *baseDevice) Addresses() []netip.Addr {
	return dev.addresses
}

func (dev *baseDevice) EOF() <-chan struct{} {
	return nil
}

// internalDevice wraps baseDevice and swaps input/output channels. This is required
// to properly forward packets between devices because the internal device's output is
// connected to the base device's input and the internal device's input is connected
// to the base device's output.
type internalDevice struct {
	*baseDevice
}

func (id *internalDevice) Input() chan<- *packet.Packet {
	return id.output
}

func (id *internalDevice) Output() <-chan *packet.Packet {
	return id.input
}

// externalDevice presents the public interface of the
// geographic link. It preserves the normal channel direction
// (input for receiving, output for sending) and is what
// we return to external callers.
type externalDevice struct {
	*baseDevice
}

func (ed *externalDevice) Input() chan<- *packet.Packet {
	return ed.input
}

func (ed *externalDevice) Output() <-chan *packet.Packet {
	return ed.output
}

// Extend creates a geographic link between the
// given device and the returned device.
//
// Internally, this creates the following link:
//
//	external <=> dev
//
// where:
//
// - dev is the device passed as argument
//
// - external is the device returned to the caller
//
// Packets flowing through this chain experience
// the configured delay in both directions.
//
// We create two goroutines for forwarding packets,
// which are closed when dev is closed.
func Extend(dev packet.NetworkDevice, config *Config) packet.NetworkDevice {
	input, output := packet.NewNetworkDeviceIOChannels()
	local := &baseDevice{
		addresses: dev.Addresses(),
		input:     input,
		output:    output,
	}
	go forward(dev, &internalDevice{local}, config)
	go forward(&internalDevice{local}, dev, config)
	return &externalDevice{local}
}

type sourceDevice interface {
	EOF() <-chan struct{}
	Output() <-chan *packet.Packet
}

type destDevice interface {
	EOF() <-chan struct{}
	Input() chan<- *packet.Packet
}

// forward implements packet forwarding with propagation delay.
//
// It maintains a queue of packets and uses a timer to implement the
// configured delay. The timer is only active when there are
// packets to forward, otherwise it runs with a long interval to
// avoid consuming resources.
//
// Packets are forwarded in order and the delay is applied to each
// packet individually. This models how packets travel through a
// physical link where the propagation delay applies to each packet.
func forward(src sourceDevice, dst destDevice, config *Config) {
	delay := max(time.Millisecond, config.Delay)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	var packets []*packet.Packet
	for {
		select {
		case pkt := <-src.Output():
			packets = append(packets, pkt)
			if len(packets) == 1 {
				ticker.Reset(delay)
			}

		case <-ticker.C:
			pkt := packets[0]
			packets = packets[1:]
			if len(packets) <= 0 {
				ticker.Reset(time.Minute)
			}

			if config.Log {
				log.Printf("geolink: %s", pkt)
			}

			select {
			case dst.Input() <- pkt:
				// delivered to destination
			case <-src.EOF():
				return
			case <-dst.EOF():
				return
			}

		case <-src.EOF():
			return
		case <-dst.EOF():
			return
		}
	}
}
