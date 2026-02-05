// SPDX-License-Identifier: GPL-3.0-or-later

package qa

import (
	"net"
	"strconv"
	"time"

	"github.com/stretchr/testify/require"
)

// MatchPattern indicates what kinds of messages an
// [*ExpectedEvent] can match.
//
// It's used to handle non-deterministic sequences of events
// that may appear zero or more times in the actual event stream.
type MatchPattern int

const (
	// MatchAnyRead matches zero or more readStart/readDone events.
	//
	// This is useful because the number of read operations might vary
	// depending on background goroutine scheduling.
	MatchAnyRead = MatchPattern(1 << iota)

	// MatchAnyWrite matches zero or more writeStart/writeDone events.
	//
	// Like reads, the number of write operations might vary depending
	// on background goroutine scheduling.
	MatchAnyWrite

	// MatchAnyClose matches zero or more closeStart/closeDone events.
	//
	// Close events might appear at different points due to connection
	// pooling and cleanup goroutine scheduling.
	MatchAnyClose
)

// ExpectedEvent describes an expected event in a measurement sequence.
type ExpectedEvent struct {
	// Msg is the expected message type. If Pattern is non-zero,
	// this field is ignored and the Pattern is used instead.
	Msg string

	// When Pattern is non-zero, this [*ExpectedEvent] acts like a
	// wildcard that consumes all matching events until the next
	// non-Pattern expectation is found.
	Pattern MatchPattern
}

// Event is an Event emitted by the RBMK tool.
type Event struct {
	//
	// Core fields
	//

	// Msg is the event identifier
	Msg string `json:"msg"`

	// T0 is the optional start timestamp for the event duration
	T0 time.Time `json:"t0"`

	// T is the event timestamp
	T time.Time `json:"t"`

	//
	// Network fields
	//

	// Protocol is the network protocol (e.g., "tcp", "udp").
	Protocol string `json:"protocol,omitempty"`

	// LocalAddr is the local endpoint address (IP:port).
	LocalAddr string `json:"localAddr,omitempty"`

	// RemoteAddr is the remote endpoint address (IP:port).
	RemoteAddr string `json:"remoteAddr,omitempty"`

	//
	// Failure
	//

	// Err is the Go error that occurred or "" on success.
	Err string `json:"err,omitempty"`

	// ErrClass is the error classification to a fixed set
	// of enumerated strings or "" on success.
	ErrClass string `json:"errClass,omitempty"`

	//
	// I/O operations
	//

	// IOBufferSize is the size of the I/O buffer.
	IOBufferSize int64 `json:"ioBufferSize,omitempty"`

	// IOBytesCount is the number of bytes read or written.
	IOBytesCount int64 `json:"ioBytesCount,omitempty"`

	//
	// DNS-specific fields
	//

	// DNSRawQuery is the raw DNS query.
	DNSRawQuery []byte `json:"dnsRawQuery,omitempty"`

	// DNSRawResponse is the raw DNS response.
	DNSRawResponse []byte `json:"dnsRawResponse,omitempty"`

	// DNSLookupDomain is the domain passed to LookupHost.
	DNSLookupDomain string `json:"dnsLookupDomain,omitempty"`

	// DNSResolvedAddrs is the list of resolved addresses.
	DNSResolvedAddrs []string `json:"dnsResolvedAddrs,omitempty"`

	//
	// Server-specific fields
	//

	// ServerAddr is the server address.
	//
	// This field is currently used for DNS but may be extended
	// to other applicative protocols in the future.
	ServerAddr string `json:"serverAddr,omitempty"`

	// ServerProtocol is the server protocol.
	//
	// This field is currently used for DNS but may be extended
	// to other applicative protocols in the future.
	ServerProtocol string `json:"serverProtocol,omitempty"`

	//
	// TLS-specific fields
	//

	// TLSServerName is the TLS server name.
	TLSServerName string `json:"tlsServerName,omitempty"`

	// TLSSkipVerify is true if the TLS verification was skipped.
	TLSSkipVerify bool `json:"tlsSkipVerify,omitempty"`

	// TLSCipherSuite is the negotiated TLS cipher suite.
	TLSCipherSuite string `json:"tlsCipherSuite,omitempty"`

	// TLSNegotiatedProto is the negotiated TLS protocol.
	TLSNegotiatedProto string `json:"tlsNegotiatedProtocol,omitempty"`

	// TLSVersion is the negotiated TLS version.
	TLSVersion string `json:"tlsVersion,omitempty"`

	// TLSPeerCerts is the list of TLS peer certificates.
	TLSPeerCerts [][]byte `json:"tlsPeerCerts,omitempty"`
}

// VerifyReadWriteClose checks that the current [*Event] matches
// the expectations for a read/write/close operation.
func (ev *Event) VerifyReadWriteClose(t Driver) {
	switch ev.Msg {
	case "readStart", "writeStart":
		ev.verifyStartEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		ev.verifyErrEmpty(t)
		ev.verifyErrClassEmpty(t)
		ev.verifyIOBufferSizePositive(t)
		ev.verifyIOBytesCountZero(t)

	case "closeStart":
		ev.verifyStartEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		ev.verifyErrEmpty(t)
		ev.verifyErrClassEmpty(t)
		ev.verifyIOBufferSizeZero(t)
		ev.verifyIOBytesCountZero(t)

	case "readDone", "writeDone":
		ev.verifyDoneEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		ev.verifyIOBufferSizeZero(t)
		ev.verifyIOBytesCountOrErr(t)
		ev.verifyIOBytesCountOrErrClass(t)

	case "closeDone":
		ev.verifyDoneEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		// any value of err is okay
		// any value of errClass is okay
		ev.verifyIOBufferSizeZero(t)
		ev.verifyIOBytesCountZero(t)

	default:
		require.Fail(t, "unexpected message %q", ev.Msg)
	}

	ev.verifyDNSRawQueryEmpty(t)
	ev.verifyDNSRawResponseEmpty(t)
	ev.verifyDNSLookupDomainEmpty(t)
	ev.verifyDNSResolverAddrsEmpty(t)
	ev.verifyServerAddrEmpty(t)
	ev.verifyServerProtocolEmpty(t)
	ev.verifyTLSServerNameEmpty(t)
	ev.verifyTLSSkipVerifyFalse(t)
	ev.verifyTLSCipherSuiteEmpty(t)
	ev.verifyTLSNegotiatedProtoEmpty(t)
	ev.verifyTLSVersionEmpty(t)
	ev.verifyTLSPeerCertsEmpty(t)
}

func (ev *Event) verifyStartEventTime(t Driver) {
	require.False(t, ev.T.IsZero(), "expected non-zero t field")
	require.True(t, ev.T0.IsZero(), "expected zero t0 field")
}

func (ev *Event) verifyDoneEventTime(t Driver) {
	require.False(t, ev.T.IsZero(), "expected non-zero t field")
	require.False(t, ev.T0.IsZero(), "expected non-zero t0 field")
	require.False(t, ev.T.Before(ev.T0), "expected t >= t0")
}

func (ev *Event) verifyProtocol(t Driver) {
	require.True(t,
		ev.Protocol == "tcp" || ev.Protocol == "udp",
		"expected protocol to be tcp or udp")
}

func (ev *Event) verifyEndpoint(t Driver, epnt string) {
	addr, port, err := net.SplitHostPort(epnt)
	require.NoError(t, err, "expected valid endpoint")
	require.True(t, net.ParseIP(addr) != nil, "expected valid IP address")
	pnum, err := strconv.Atoi(port)
	require.NoError(t, err, "expected valid port number")
	require.True(t, pnum >= 1 && pnum <= 65535, "expected valid port number")
}

func (ev *Event) verifyErrEmpty(t Driver) {
	require.Empty(t, ev.Err, "expected empty error field")
}

func (ev *Event) verifyErrClassEmpty(t Driver) {
	require.Empty(t, ev.ErrClass, "expected empty errClass field")
}

func (ev *Event) verifyIOBufferSizePositive(t Driver) {
	require.True(t, ev.IOBufferSize > 0, "expected positive ioBufferSize field")
}

func (ev *Event) verifyIOBufferSizeZero(t Driver) {
	require.Zero(t, ev.IOBufferSize, "expected zero ioBufferSize field")
}

func (ev *Event) verifyIOBytesCountZero(t Driver) {
	require.Zero(t, ev.IOBytesCount, "expected zero ioBytesCount field")
}

func (ev *Event) verifyIOBytesCountOrErr(t Driver) {
	require.True(t, ev.IOBytesCount > 0 || ev.Err != "", "expected ioBytesCount > 0 or err != \"\"")
}

func (ev *Event) verifyIOBytesCountOrErrClass(t Driver) {
	require.True(t, ev.IOBytesCount > 0 || ev.ErrClass != "", "expected ioBytesCount > 0 or errClass != \"\"")
}

func (ev *Event) verifyDNSRawQueryEmpty(t Driver) {
	require.Empty(t, ev.DNSRawQuery, "expected empty dnsRawQuery field")
}

func (ev *Event) verifyDNSRawResponseEmpty(t Driver) {
	require.Empty(t, ev.DNSRawResponse, "expected empty dnsRawResponse field")
}

func (ev *Event) verifyDNSLookupDomainEmpty(t Driver) {
	require.Empty(t, ev.DNSLookupDomain, "expected empty dnsLookupDomain field")
}

func (ev *Event) verifyDNSResolverAddrsEmpty(t Driver) {
	require.Empty(t, ev.DNSResolvedAddrs, "expected empty dnsResolvedAddrs field")
}

func (ev *Event) verifyServerAddrEmpty(t Driver) {
	require.Empty(t, ev.ServerAddr, "expected empty serverAddr field")
}

func (ev *Event) verifyServerProtocolEmpty(t Driver) {
	require.Empty(t, ev.ServerProtocol, "expected empty serverProtocol field")
}

func (ev *Event) verifyTLSServerNameEmpty(t Driver) {
	require.Empty(t, ev.TLSServerName, "expected empty tlsServerName field")
}

func (ev *Event) verifyTLSSkipVerifyFalse(t Driver) {
	require.False(t, ev.TLSSkipVerify, "expected false tlsSkipVerify field")
}

func (ev *Event) verifyTLSCipherSuiteEmpty(t Driver) {
	require.Empty(t, ev.TLSCipherSuite, "expected empty tlsCipherSuite field")
}

func (ev *Event) verifyTLSNegotiatedProtoEmpty(t Driver) {
	require.Empty(t, ev.TLSNegotiatedProto, "expected empty tlsNegotiatedProtocol field")
}

func (ev *Event) verifyTLSVersionEmpty(t Driver) {
	require.Empty(t, ev.TLSVersion, "expected empty tlsVersion field")
}

func (ev *Event) verifyTLSPeerCertsEmpty(t Driver) {
	require.Empty(t, ev.TLSPeerCerts, "expected empty tlsPeerCerts field")
}

// VerifyEqual checks whether an event is equal to another.
func (expect *ExpectedEvent) VerifyEqual(t Driver, got *Event) {
	require.Equal(t, expect.Msg, got.Msg, "expected %q, got %q", expect.Msg, got.Msg)
}
