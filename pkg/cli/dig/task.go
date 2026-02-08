// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/bassosimone/dnscodec"
	"github.com/miekg/dns"
	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/rbmk-project/rbmk/internal/testablenet"
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
var protocolMap = map[string]struct{}{
	"udp": {},
	"tcp": {},
	"dot": {},
	"doh": {},
}

// newServerAddr returns a new server address string based on the protocol
// and the specific fields configured for the task.
func (task *Task) newServerAddr() *url.URL {
	switch task.Protocol {
	case "udp", "tcp", "dot":
		return &url.URL{
			Scheme: task.Protocol,
			Host:   net.JoinHostPort(task.ServerAddr, task.ServerPort),
			Path:   "/",
		}

	case "doh":
		return &url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(task.ServerAddr, task.ServerPort),
			Path:   task.URLPath,
		}

	default:
		panic(fmt.Errorf("unsupported protocol: %s", task.Protocol))
	}
}

// Run runs the task and returns an error.
func (task *Task) Run(ctx context.Context) error {
	// Setup the overall operation timeout using the context
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set up the JSON logger for writing the measurements
	logger := slog.New(task.newSlogHandler())

	// Create netcore network instance
	netx := netcore.NewNetwork()
	netx.DialContextFunc = testablenet.DialContext.Get()
	netx.Logger = logger

	// Determine the DNS query type
	queryType, ok := queryTypeMap[task.QueryType]
	if !ok {
		return fmt.Errorf("unsupported query type: %s", task.QueryType)
	}

	// Validate the server protocol
	if _, ok := protocolMap[task.Protocol]; !ok {
		return fmt.Errorf("unsupported protocol: %s", task.Protocol)
	}

	// Create the server address
	URL := task.newServerAddr()

	// Create the DNS query
	query := dnscodec.NewQuery(task.Name, queryType)

	// Handle the `+udp=wait-duplicates` special case
	if URL.Scheme == "udp" && task.WaitDuplicates {
		return task.waitUDP(ctx, logger, netx, URL, query) // error already wrapped
	}

	// Connect to the server
	conn, err := netx.DialDNS(ctx, URL)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	// Perform the exchange; the response is intentionally discarded because
	// the slogHandler intercepts raw DNS messages from structured logs and
	// prints them to the configured writers.
	if _, err := conn.Exchange(ctx, query); err != nil {
		return fmt.Errorf("exchange failed: %w", err)
	}
	return nil
}
