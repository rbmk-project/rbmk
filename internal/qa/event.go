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
	T0 time.Time `json:"t0,omitempty"`

	// T is the event timestamp
	T time.Time `json:"t,omitempty"`

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

	// Err is the Go error that occurred.
	Err string `json:"err,omitempty"`

	//
	// I/O operations
	//

	// Count is the number of bytes read or written.
	Count int64 `json:"count,omitempty"`

	//
	// DNS-specific fields
	//

	// RawQuery is the raw DNS query message encoded as base64.
	RawQuery []byte `json:"rawQuery,omitempty"`

	// RawResponse is the raw DNS response message encoded as base64.
	RawResponse []byte `json:"rawResponse,omitempty"`

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
		ev.verifyCountPositive(t)

	case "closeStart":
		ev.verifyStartEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		ev.verifyErrEmpty(t)
		ev.verifyCountZero(t)

	case "readDone", "writeDone":
		ev.verifyDoneEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		ev.verifyCountOrErr(t)

	case "closeDone":
		ev.verifyDoneEventTime(t)
		ev.verifyProtocol(t)
		ev.verifyEndpoint(t, ev.LocalAddr)
		ev.verifyEndpoint(t, ev.RemoteAddr)
		// any value of error is okay
		ev.verifyCountZero(t)

	default:
		require.Fail(t, "unexpected message %q", ev.Msg)
	}

	ev.verifyRawQueryEmpty(t)
	ev.verifyRawResponseEmpty(t)
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

func (ev *Event) verifyCountPositive(t Driver) {
	require.True(t, ev.Count > 0, "expected positive count field")
}

func (ev *Event) verifyCountZero(t Driver) {
	require.Zero(t, ev.Count, "expected zero count field")
}

func (ev *Event) verifyCountOrErr(t Driver) {
	require.True(t, ev.Count > 0 || ev.Err != "", "expected count or error")
}

func (ev *Event) verifyRawQueryEmpty(t Driver) {
	require.Empty(t, ev.RawQuery, "expected empty rawQuery field")
}

func (ev *Event) verifyRawResponseEmpty(t Driver) {
	require.Empty(t, ev.RawResponse, "expected empty rawResponse field")
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
	// Make sure the messages are equal
	require.Equal(t, expect.Msg, got.Msg, "expected %q, got %q", expect.Msg, got.Msg)
}
