// SPDX-License-Identifier: GPL-3.0-or-later

package qa

import "github.com/stretchr/testify/require"

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

	// Pattern indicates if this expectation matches zero or more
	// events of specific types (e.g., reads, writes, closes).
	//
	// When Pattern is non-zero, this [*ExpectedEvent] acts like a
	// wildcard that consumes all matching events until the next
	// non-Pattern expectation is found.
	Pattern MatchPattern
}

// Event is an Event emitted by the RBMK tool.
type Event struct {
	Msg string `json:"msg"`
}

// VerifyEqual checks whether an event is equal to another.
func (expect *ExpectedEvent) VerifyEqual(t Driver, got *Event) {
	// Make sure the messages are equal
	require.Equal(t, expect.Msg, got.Msg, "expected %q, got %q", expect.Msg, got.Msg)
}
