// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package testable provides thread-safe singletons for overriding
fundamental RBMK dependencies in integration tests.

The zero value of each singleton is ready to use and typically
uses the standard library. Overriding to a different value allows
to either use mocks or replacements such as the ones implemented
by the rbmk-project/x/netsim package.
*/
package testable

import (
	"context"
	"crypto/x509"
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

// RootCAsProvider provides a thread-safe way to override the root CAs.
//
// The zero value is ready to use and uses the system root CAs.
type RootCAsProvider struct {
	pool *x509.CertPool
	mu   sync.Mutex
}

// RootCAs is the singleton allowing to override the root CAs.
//
// By default, we use the system root CAs.
var RootCAs = &RootCAsProvider{}

// Set sets the RootCA pool to use.
func (rcp *RootCAsProvider) Set(pool *x509.CertPool) {
	rcp.mu.Lock()
	defer rcp.mu.Unlock()
	rcp.pool = pool
}

// Get returns the RootCA pool to use.
func (rcp *RootCAsProvider) Get() *x509.CertPool {
	rcp.mu.Lock()
	defer rcp.mu.Unlock()
	return rcp.pool
}
