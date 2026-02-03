// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/nop"
	"github.com/rbmk-project/rbmk/pkg/cli/internal/testable"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
)

// exchangeUDP performs an exchange over UDP.
func (task *Task) exchangeUDP(
	ctx context.Context,
	cp *closepool.Pool,
	slogger *slog.Logger,
	serverAddr netip.AddrPort,
	query *dnscodec.Query,
) error {
	cfg := task.newNopConfig()
	pipe := nop.Compose5(
		nop.NewEndpointFunc(serverAddr),
		nop.NewConnectFunc(cfg, "udp", slogger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(cfg, slogger),
		nop.NewDNSOverUDPConnFunc(cfg, slogger),
	)
	return exchangeCommon(ctx, cp, pipe, query)
}

// exchangeTCP performs an exchange over TCP.
func (task *Task) exchangeTCP(
	ctx context.Context,
	cp *closepool.Pool,
	slogger *slog.Logger,
	serverAddr netip.AddrPort,
	query *dnscodec.Query,
) error {
	cfg := task.newNopConfig()
	pipe := nop.Compose5(
		nop.NewEndpointFunc(serverAddr),
		nop.NewConnectFunc(cfg, "tcp", slogger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(cfg, slogger),
		nop.NewDNSOverTCPConnFunc(cfg, slogger),
	)
	return exchangeCommon(ctx, cp, pipe, query)
}

// exchangeTLS performs an exchange over TLS.
func (task *Task) exchangeTLS(
	ctx context.Context,
	cp *closepool.Pool,
	slogger *slog.Logger,
	serverAddr netip.AddrPort,
	query *dnscodec.Query,
) error {
	cfg := task.newNopConfig()
	tlsConfig := &tls.Config{
		ServerName: task.ServerAddr,
		NextProtos: []string{"dot"},
		RootCAs:    testable.RootCAs.Get(),
	}
	pipe := nop.Compose6(
		nop.NewEndpointFunc(serverAddr),
		nop.NewConnectFunc(cfg, "tcp", slogger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(cfg, slogger),
		nop.NewTLSHandshakeFunc(cfg, tlsConfig, slogger),
		nop.NewDNSOverTLSConnFunc(cfg, slogger),
	)
	return exchangeCommon(ctx, cp, pipe, query)
}

// exchangeHTTPS performs an exchange over HTTPS.
func (task *Task) exchangeHTTPS(
	ctx context.Context,
	cp *closepool.Pool,
	slogger *slog.Logger,
	serverAddr netip.AddrPort,
	query *dnscodec.Query,
) error {
	cfg := task.newNopConfig()
	tlsConfig := &tls.Config{
		ServerName: task.ServerAddr,
		NextProtos: []string{"h2", "http/1.1"},
		RootCAs:    testable.RootCAs.Get(),
	}
	serverURL := &url.URL{
		Scheme: "https",
		Host:   task.ServerAddr,
		Path:   task.URLPath,
	}
	pipe := nop.Compose7(
		nop.NewEndpointFunc(serverAddr),
		nop.NewConnectFunc(cfg, "tcp", slogger),
		nop.NewCancelWatchFunc(),
		nop.NewObserveConnFunc(cfg, slogger),
		nop.NewTLSHandshakeFunc(cfg, tlsConfig, slogger),
		nop.NewHTTPConnFuncTLS(cfg, slogger),
		nop.NewDNSOverHTTPSConnFunc(cfg, serverURL.String(), slogger),
	)
	return exchangeCommon(ctx, cp, pipe, query)
}

// exchangeCloser abstracts over a DNS conn.
type exchangeCloser interface {
	Exchange(ctx context.Context, query *dnscodec.Query) (*dnscodec.Response, error)
	Close() error
}

// exchangeCommon is the common code to perform an exchange.
func exchangeCommon[T exchangeCloser](
	ctx context.Context,
	cp *closepool.Pool,
	pipe nop.Func[nop.Unit, T],
	query *dnscodec.Query,
) error {
	conn, err := pipe.Call(ctx, nop.Unit{})
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	cp.Add(conn)
	if _, err := conn.Exchange(ctx, query); err != nil {
		return fmt.Errorf("exchange failed: %w", err)
	}
	return nil
}
