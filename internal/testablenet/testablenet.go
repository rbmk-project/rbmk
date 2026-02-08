// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package testablenet provides thread-safe singletons for overriding
the fundamental network dependencies in integration tests.

The zero value of each singleton is ready to use and typically
uses the standard library. Overriding to a different value allows
to use mocks, stubs, or alternative implementations.
*/
package testablenet

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
	mu sync.RWMutex
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
	dcp.mu.RLock()
	defer dcp.mu.RUnlock()
	fx := dcp.fx
	if fx == nil {
		dialer := &net.Dialer{}
		dialer.SetMultipathTCP(false)
		fx = dialer.DialContext
	}
	return fx
}

// LookupHostFunc is the type of the low-level lookup-host function.
type LookupHostFunc func(ctx context.Context, domain string) ([]string, error)

// LookupHostProvider provides a thread-safe way to override the lookup-host function.
//
// The zero value is ready to use and resolves with the standard library.
type LookupHostProvider struct {
	fx LookupHostFunc
	mu sync.RWMutex
}

// LookupHost is the singleton allowing to override the function used
// to resolve domain names without data races.
//
// By default, we use the standard library to resolve domain names.
var LookupHost = &LookupHostProvider{}

// Set sets the lookup-host function to use to resolve domain names.
func (lhp *LookupHostProvider) Set(fx LookupHostFunc) {
	lhp.mu.Lock()
	defer lhp.mu.Unlock()
	lhp.fx = fx
}

// Get returns the lookup-host function to use to resolve domain names.
func (lhp *LookupHostProvider) Get() LookupHostFunc {
	lhp.mu.RLock()
	defer lhp.mu.RUnlock()
	fx := lhp.fx
	if fx == nil {
		fx = (&net.Resolver{}).LookupHost
	}
	return fx
}

// RootCAsProvider provides a thread-safe way to override the root CAs.
//
// The zero value is ready to use and uses the system root CAs.
type RootCAsProvider struct {
	pool *x509.CertPool
	mu   sync.RWMutex
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
	rcp.mu.RLock()
	defer rcp.mu.RUnlock()
	return rcp.pool
}
