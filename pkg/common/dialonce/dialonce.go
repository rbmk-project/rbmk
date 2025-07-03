// SPDX-License-Identifier: GPL-3.0-or-later

// Package dialonce provides a way to ensure we dial just once.
package dialonce

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
)

// DialContextFunc is the function that dials a network connection with the given network and address.
type DialContextFunc = func(ctx context.Context, network, address string) (net.Conn, error)

// singleDialer ensures we dial just once.
//
// The zero value is ready to use.
type singleDialer struct {
	count atomic.Int32
	dial  DialContextFunc
}

// ErrMultipleDial is the error returned when we dial more than once.
var ErrMultipleDial = errors.New("dialing more than once")

// DialContext dials a network connection with the given network and address.
func (d *singleDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.count.Add(1) > 1 {
		return nil, ErrMultipleDial
	}
	return d.dial(ctx, network, address)
}

// Wrap wraps a [DialContextFunc] to ensure we dial just once.
//
// Multiple attempts to dial will return [ErrMultipleDial].
func Wrap(dial DialContextFunc) DialContextFunc {
	return (&singleDialer{count: atomic.Int32{}, dial: dial}).DialContext
}
