// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"net/url"
	"os"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/minest"
	"github.com/bassosimone/nop"
	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/safeconn"
	"github.com/rbmk-project/rbmk/internal/netcore"
)

// unusedDialer is a [minest.NetDialer] that panics if DialContext is called.
//
// The waitUDP function uses a pre-established connection and never dials.
// This type serves as a sentinel to catch programming errors where the
// transport attempts to dial instead of using the provided connection.
type unusedDialer struct{}

var _ minest.NetDialer = unusedDialer{}

// DialContext implements [minest.NetDialer] and always panics.
func (unusedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	panic("dig: DNS transport must not dial; this is a programming error")
}

// waitUDP sends a query over UDP and waits for duplicate responses
// until the timeout configured in the context expires.
func (task *Task) waitUDP(
	ctx context.Context,
	slogger *slog.Logger,
	netx *netcore.Network,
	URL *url.URL,
	query *dnscodec.Query,
) error {
	// 0. Make sure there are no programming errors.
	runtimex.Assert(URL.Scheme == "udp")

	// 1. Build a 4-step pipeline to obtain a net.Conn with
	// logging, cancel-watching, and I/O observation.
	conn, err := netx.DialDNS(ctx, URL)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()
	uconn := conn.(*nop.DNSOverUDPConn).Conn() // netx.DialDNS should dial this type

	// 2. Create the log context for structured exchange logging.
	cfg := netx.NewNopConfig()
	t0 := cfg.TimeNow()
	deadline, _ := ctx.Deadline()
	var rqr []byte
	lc := &nop.DNSExchangeLogContext{
		ErrClassifier:  cfg.ErrClassifier,
		LocalAddr:      safeconn.LocalAddr(uconn),
		Logger:         slogger,
		Protocol:       safeconn.Network(uconn),
		RemoteAddr:     safeconn.RemoteAddr(uconn),
		ServerProtocol: "udp",
		TimeNow:        cfg.TimeNow,
	}

	// 3. Create the transport with observer hooks but no real dialer.
	txp := minest.NewDNSOverUDPTransport(unusedDialer{}, netip.AddrPortFrom(netip.IPv4Unspecified(), 0))
	txp.ObserveRawQuery = lc.MakeQueryObserver(t0, &rqr)
	txp.ObserveRawResponse = lc.MakeResponseObserver(t0, &rqr)

	// 4. Log exchange start and send the query once.
	lc.LogStart(t0, deadline)
	queryMsg, err := txp.SendQuery(ctx, uconn, query)
	if err != nil {
		lc.LogDone(t0, deadline, err)
		return fmt.Errorf("send query failed: %w", err)
	}

	// 5. Loop receiving responses until the context deadline expires
	// or an unexpected error occurs. Each response triggers the
	// ObserveRawResponse hook, which emits a structured log that the
	// slogHandler intercepts and prints to stdout.
	var count uint64
	for {
		if _, err = txp.RecvResponse(ctx, uconn, queryMsg); err != nil {
			// Remap common errors to nil when we received 1+ DNS responses
			switch {
			case count > 0 && ctx.Err() != nil:
				err = nil
			case count > 0 && errors.Is(err, os.ErrDeadlineExceeded):
				err = nil
			case count > 0 && errors.Is(err, net.ErrClosed):
				err = nil
			}
			lc.LogDone(t0, deadline, err)
			if err != nil {
				return fmt.Errorf("recv response failed: %w", err)
			}
			return nil
		}
		count++
	}
}
