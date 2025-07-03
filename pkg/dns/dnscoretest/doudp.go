// SPDX-License-Identifier: GPL-3.0-or-later

package dnscoretest

import (
	"net"

	"github.com/rbmk-project/rbmk/pkg/common/runtimex"
)

// StartUDP starts an UDP listener and listens for incoming DNS queries.
//
// This method panics in case of failure.
func (s *Server) StartUDP(handler Handler) <-chan struct{} {
	runtimex.Assert(!s.started, "already started")
	ready := make(chan struct{})
	go func() {
		pconn := runtimex.Try1(s.listenPacket("udp", "127.0.0.1:0"))
		s.Addr = pconn.LocalAddr().String()
		s.ioclosers = append(s.ioclosers, pconn)
		s.started = true
		close(ready)
		for s.servePacketConn(handler, pconn) == nil {
			// nothing
		}
	}()
	return ready
}

// listenPacket either uses the standard library or the custom ListenPacket func.
func (s *Server) listenPacket(network, address string) (net.PacketConn, error) {
	if s.ListenPacket != nil {
		return s.ListenPacket(network, address)
	}
	return net.ListenPacket(network, address)
}

// servePacketConn serves a single DNS query over UDP.
func (s *Server) servePacketConn(handler Handler, pconn net.PacketConn) error {
	buf := make([]byte, 4096)
	count, addr, err := pconn.ReadFrom(buf)
	if err != nil {
		return err
	}
	rawQuery := buf[:count]
	rw := &responseWriterUDP{pconn: pconn, addr: addr}
	handler.Handle(rw, rawQuery)
	return nil
}

// responseWriterUDP is a response writer for UDP.
type responseWriterUDP struct {
	pconn net.PacketConn
	addr  net.Addr
}

// Ensure responseWriterUDP implements ResponseWriter.
var _ ResponseWriter = (*responseWriterUDP)(nil)

// Write implements ResponseWriter.
func (r *responseWriterUDP) Write(rawMsg []byte) (int, error) {
	return r.pconn.WriteTo(rawMsg, r.addr)
}
