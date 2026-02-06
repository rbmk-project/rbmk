// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/bassosimone/nop"
	"github.com/miekg/dns"
	"github.com/rbmk-project/rbmk/pkg/cli/internal/testable"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
	"github.com/rbmk-project/rbmk/pkg/common/errclass"
)

// Task runs the `dig` task.
//
// The zero value is not ready to use. Please, make sure
// to initialize all the fields marked as MANDATORY.
type Task struct {
	// LogsWriter is the MANDATORY [io.Writer] where
	// we should write structured logs.
	LogsWriter io.Writer

	// Name is the MANDATORY name to query.
	Name string

	// Protocol is the MANDATORY protocol to use,
	// expressed as a string. For example, "udp" or "tcp".
	//
	// See [dnscore.NewServerAddr] for more details.
	Protocol string

	// QueryType is the MANDATORY query type expressed
	// as a string. For example, "A" or "AAAA".
	QueryType string

	// QueryWriter is the MANDATORY [io.Writer] where we should
	// write the query before sending it.
	QueryWriter io.Writer

	// ResponseWriter is the MANDATORY [io.Writer] where we should
	// write the full response when we received it.
	ResponseWriter io.Writer

	// ShortIP is a flag that ensures that `+short=ip` only
	// prints the IP addresses in the response.
	ShortIP bool

	// ShortWriter is the MANDATORY [io.Writer] where we should
	// write the short response when we received it.
	ShortWriter io.Writer

	// ServerAddr is the MANDATORY address of the server
	// to query, for example "8.8.8.8", "1.1.1.1".
	ServerAddr string

	// ServerPort is the MANDATORY port of the server to
	// query. For example, "53".
	ServerPort string

	// URLPath is the MANDATORY URL path when using DoH.
	URLPath string

	// WaitDuplicates is the OPTIONAL flag indicating
	// whether we should wait for duplicate DNS-over-UDP
	// responses (for detecting censorship).
	WaitDuplicates bool
}

// queryTypeMap maps query types strings to DNS query types.
var queryTypeMap = map[string]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
	"HTTPS": dns.TypeHTTPS,
	"MX":    dns.TypeMX,
	"NS":    dns.TypeNS,
}

// Run runs the task and returns an error.
func (task *Task) Run(ctx context.Context) error {
	// Setup the overal operation timeout using the context
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set up the JSON logger for writing the measurements
	logger := slog.New(task.newSlogHandler())
	logger = logger.With("spanID", nop.NewSpanID())

	// Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// Create the query
	queryType, ok := queryTypeMap[task.QueryType]
	if !ok {
		return fmt.Errorf("unsupported query type: %s", task.QueryType)
	}
	query := dnscodec.NewQuery(task.Name, queryType)

	// TODO(bassosimone): before refactoring to `nop` we supported `@dns.google`
	// as an argument, while now that support is gone and this is bad

	// Create the endpoint address
	addr, err := netip.ParseAddrPort(net.JoinHostPort(task.ServerAddr, task.ServerPort))
	if err != nil {
		return fmt.Errorf("cannot parse server endpoint: %w", err)
	}

	// Dispatch control depending on protocol and settings.
	switch {
	case task.Protocol == "udp" && !task.WaitDuplicates:
		return task.exchangeUDP(ctx, pool, logger, addr, query)

	case task.Protocol == "udp" && task.WaitDuplicates:
		return task.waitUDP(ctx, pool, logger, addr, query)

	case task.Protocol == "tcp":
		return task.exchangeTCP(ctx, pool, logger, addr, query)

	case task.Protocol == "dot":
		return task.exchangeTLS(ctx, pool, logger, addr, query)

	case task.Protocol == "doh":
		return task.exchangeHTTPS(ctx, pool, logger, addr, query)

	default:
		return fmt.Errorf("unsupported protocol type: %s", task.Protocol)
	}
}

// dialerContextFunc adapts a dial-context-like function to become a [nop.Dialer].
type dialerContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

var _ nop.Dialer = dialerContextFunc(nil)

// DialContext implements [nop.Dialer].
func (fx dialerContextFunc) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	return fx(ctx, network, address)
}

// newNopConfig creates a new [*nop.Config] instance.
func (task *Task) newNopConfig() *nop.Config {
	cfg := nop.NewConfig()
	cfg.ErrClassifier = nop.ErrClassifierFunc(errclass.New)
	cfg.Dialer = dialerContextFunc(testable.DialContext.Get())
	return cfg
}
