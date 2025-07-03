//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/runtimex/runtimex.go
//

// Package runtimex contains runtime extensions.
//
// This package is inspired to https://pkg.go.dev/github.com/m-lab/go/rtx, except that it's simpler.
package runtimex

import (
	"errors"
	"fmt"
)

// PanicOnError calls panic() if err is not nil. The type passed
// to panic is an error type wrapping the original error.
func PanicOnError(err error, message string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", message, err))
	}
}

// Assert calls panic if assertion is false. The type passed to
// panic is an error constructed using errors.New(message).
func Assert(assertion bool, message string) {
	if !assertion {
		panic(errors.New(message))
	}
}

// Try0 calls [runtimex.PanicOnError] if err is not nil.
func Try0(err error) {
	PanicOnError(err, "Try0")
}

// Try1 is like [Try0] but supports functions returning one values and an error.
func Try1[T1 any](v1 T1, err error) T1 {
	PanicOnError(err, "Try1")
	return v1
}

// Try2 is like [Try1] but supports functions returning two values and an error.
func Try2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	PanicOnError(err, "Try2")
	return v1, v2
}

// Try3 is like [Try2] but supports functions returning three values and an error.
func Try3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3) {
	PanicOnError(err, "Try3")
	return v1, v2, v3
}
