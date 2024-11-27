// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/miekg/dns"
	"github.com/rbmk-project/dnscore"
	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/rbmk-project/x/closepool"
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
	// Set up the JSON logger for writing the measurements
	logger := slog.New(slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{}))

	// Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// Create netcore network instance
	netx := &netcore.Network{}
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
	response, err := transport.Query(context.Background(), server, query)
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
			fmt.Fprintf(&builder, "%s\n", ans.Target)

		case *dns.HTTPS:
			value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
			fmt.Fprintf(&builder, "%s\n", value)

		case *dns.MX:
			value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
			fmt.Fprintf(&builder, "%s\n", value)

		case *dns.NS:
			value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
			fmt.Fprintf(&builder, "%s\n", value)

		default:
			// TODO(bassosimone): implement the other answer types
		}
	}
	return builder.String()
}
