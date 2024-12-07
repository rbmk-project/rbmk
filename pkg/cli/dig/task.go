// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/rbmk-project/common/closepool"
	"github.com/rbmk-project/dnscore"
	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/rbmk-project/x/netcore"
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

// protocolMap maps protocol strings to DNS protocols.
var protocolMap = map[string]dnscore.Protocol{
	"udp": dnscore.ProtocolUDP,
	"tcp": dnscore.ProtocolTCP,
	"dot": dnscore.ProtocolDoT,
	"doh": dnscore.ProtocolDoH,
}

// newServerAddr returns a new server address string based on the protocol
// and the specific fields configured for the task.
func (task *Task) newServerAddr(protocol dnscore.Protocol) string {
	switch protocol {
	case dnscore.ProtocolUDP, dnscore.ProtocolTCP, dnscore.ProtocolDoT:
		return net.JoinHostPort(task.ServerAddr, task.ServerPort)

	case dnscore.ProtocolDoH:
		URL := &url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(task.ServerAddr, task.ServerPort),
			Path:   task.URLPath,
		}
		return URL.String()

	default:
		panic(fmt.Errorf("unsupported protocol: %s", protocol))
	}
}

// Run runs the task and returns an error.
func (task *Task) Run(ctx context.Context) error {
	// Setup the overal operation timeout using the context
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set up the JSON logger for writing the measurements
	logger := slog.New(slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{}))

	// Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// Create netcore network instance
	netx := &netcore.Network{}
	netx.RootCAs = testable.RootCAs.Get()
	netx.DialContextFunc = testable.DialContext.Get()
	netx.Logger = logger
	netx.WrapConn = func(ctx context.Context, netx *netcore.Network, conn net.Conn) net.Conn {
		conn = netcore.WrapConn(ctx, netx, conn)
		pool.Add(conn)
		return conn
	}

	// Create a new transport using the logger and the network
	transport := &dnscore.Transport{}
	transport.DialContext = netx.DialContext
	transport.DialTLSContext = netx.DialTLSContext
	transport.HTTPClient = &http.Client{
		Timeout: timeout, // ensure the overall operation is bounded
		Transport: &http.Transport{
			DialContext:       netx.DialContext,
			DialTLSContext:    netx.DialTLSContext,
			ForceAttemptHTTP2: true,
		},
	}
	transport.Logger = logger

	// Determine the DNS query type
	queryType, ok := queryTypeMap[task.QueryType]
	if !ok {
		return fmt.Errorf("unsupported query type: %s", task.QueryType)
	}

	// Determine the server protocol
	protocol, ok := protocolMap[task.Protocol]
	if !ok {
		return fmt.Errorf("unsupported protocol: %s", task.Protocol)
	}

	// Create the server address
	server := dnscore.NewServerAddr(protocol, task.newServerAddr(protocol))
	flags := 0
	maxlength := uint16(dnscore.EDNS0SuggestedMaxResponseSizeUDP)
	if protocol == dnscore.ProtocolDoT || protocol == dnscore.ProtocolDoH {
		flags |= dnscore.EDNS0FlagDO | dnscore.EDNS0FlagBlockLengthPadding
	}
	if protocol != dnscore.ProtocolUDP {
		maxlength = dnscore.EDNS0SuggestedMaxResponseSizeOtherwise
	}

	// Create the DNS query
	optEDNS0 := dnscore.QueryOptionEDNS0(maxlength, flags)
	query, err := dnscore.NewQuery(task.Name, queryType, optEDNS0)
	if err != nil {
		return fmt.Errorf("cannot create query: %w", err)
	}
	fmt.Fprintf(task.QueryWriter, ";; Query:\n%s\n", query.String())

	// Perform the DNS query
	response, err := task.query(ctx, transport, server, query)
	if err != nil {
		return fmt.Errorf("query round-trip failed: %w", err)
	}
	fmt.Fprintf(task.ResponseWriter, "\n;; Response:\n%s\n\n", response.String())
	fmt.Fprintf(task.ShortWriter, "%s", task.formatShort(response))

	// Explicitly close the connections in the pool
	pool.Close()

	// Validate the DNS response
	if err = dnscore.ValidateResponse(query, response); err != nil {
		return fmt.Errorf("cannot validate response: %w", err)
	}

	// Map the RCODE to an error, if any
	if err := dnscore.RCodeToError(response); err != nil {
		return fmt.Errorf("response code indicates error: %w", err)
	}
	return nil
}

// query performs the query and returns response or error.
//
// If the WaitDuplicates flag is set, this function will wait
// for duplicate responses, emit all the related structured logs,
// and return the first response received. This function blocks
// until the timeout configured in the context expires. Note that
// all responses (including duplicates) are automatically
// logged through the transport's logger.
func (task *Task) query(
	ctx context.Context,
	txp *dnscore.Transport,
	addr *dnscore.ServerAddr,
	query *dns.Msg,
) (*dns.Msg, error) {
	// If we're not waiting for duplicates, our job is easy
	if !task.WaitDuplicates {
		return txp.Query(ctx, addr, query)
	}

	// Otherwise, we need to reading duplicate responses
	// until the overall timeout says we should bail, which
	// happens through context expiration.
	var (
		resp0 *dns.Msg
		err0  error
		once  sync.Once
	)
	respch := txp.QueryWithDuplicates(ctx, addr, query)
	for entry := range respch {
		resp, err := entry.Msg, entry.Err
		once.Do(func() {
			resp0, err0 = resp, err
		})
	}
	if resp0 == nil && err0 == nil {
		return nil, errors.New("received nil response and nil error")
	}
	return resp0, err0
}

// formatShort returns a short string representation of the DNS response.
func (task *Task) formatShort(response *dns.Msg) string {
	var builder strings.Builder
	for _, ans := range response.Answer {
		switch ans := ans.(type) {
		case *dns.A:
			fmt.Fprintf(&builder, "%s\n", ans.A.String())

		case *dns.AAAA:
			fmt.Fprintf(&builder, "%s\n", ans.AAAA.String())

		case *dns.CNAME:
			if !task.ShortIP {
				fmt.Fprintf(&builder, "%s\n", ans.Target)
			}

		case *dns.HTTPS:
			if !task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		case *dns.MX:
			if !task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		case *dns.NS:
			if !task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		default:
			// TODO(bassosimone): implement the other answer types
		}
	}
	return builder.String()
}
