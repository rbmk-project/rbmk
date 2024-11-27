// SPDX-License-Identifier: GPL-3.0-or-later

package testable

import (
	"context"
	"net"
	"sync"
)

// DialContextFunc is the type of the low-level dial function.
type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

// DialContextProvider provides a thread-safe way to override the dial function.
//
// The zero value is ready to use and dials with the standard library.
type DialContextProvider struct {
	fx DialContextFunc
	mu sync.Mutex
}

// DialContext is the singleton allowing to override the function used
// to establish network connections without data races.
//
// By default, we use the standard library to dial connections.
var DialContext = &DialContextProvider{}

// Set sets the dial function to use to establish a new network connection.
func (dcp *DialContextProvider) Set(fx DialContextFunc) {
	dcp.mu.Lock()
	defer dcp.mu.Unlock()
	dcp.fx = fx
}

// Get returns the dial function to use to establish a new network connection.
func (dcp *DialContextProvider) Get() DialContextFunc {
	dcp.mu.Lock()
	defer dcp.mu.Unlock()
	return dcp.fx
}
