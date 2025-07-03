// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package httpconntrace provides a way to trace the local and remote endpoints
used by an HTTP connection while performing an [*http.Client] request.

Internally, we use [net/http/httptrace] to collect the connection [*Endpoints].

Operationally, you need to use [Do] where you would otherwise call
[*http.Client.Do] method. The [*Endpoints] are returned along with the response.

Collecting the connection [*Endpoints] is important to map the HTTP response
with the connection that actually serviced the request.
*/
package httpconntrace

import (
	"context"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/netip"
	"sync"
)

// Endpoints contains the connection endpoints extacted by [Do].
type Endpoints struct {
	// LocalAddr is the local address of the connection.
	LocalAddr netip.AddrPort

	// RemoteAddr is the remote address of the connection.
	RemoteAddr netip.AddrPort
}

// Do performs an HTTP request using [*http.Client.Do] and uses [net/http/httptrace] to
// extract the local and remote [*Endpoints] used by the connection.
//
// Internally, this function creates a new context for tracing purposes, to avoid
// accidentally composing the [net/http/httptrace] trace with other possible context traces
// that may have already been present in the request context. Obviously, this means that
// using this function prevents one to observe connection events with a trace.
//
// Note that this function assumes we're using TCP and casts the connection addresses
// to [*net.TCPAddr] to extract the endpoints. If the we're not using TCP, the returned
// [*Endpoint] will contain zero initialized (i.e., invalid) addresses.
//
// We return *Endpoints rather than Endpoints because the structure is larger than 32 bytes
// and could possibly be further extended in the future to include additional fields.
func Do(client *http.Client, req *http.Request) (*http.Response, *Endpoints, error) {
	// Prepare to collect info in a goroutine-safe way.
	var (
		laddr netip.AddrPort
		mu    sync.Mutex
		raddr netip.AddrPort
	)

	// Create clean context for tracing where "clean" means
	// we don't compose with other possible context traces
	traceCtx, cancel := context.WithCancel(context.Background())

	// Configure the trace for extracting laddr, raddr
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			mu.Lock()
			defer mu.Unlock()
			if addr, ok := info.Conn.LocalAddr().(*net.TCPAddr); ok {
				laddr = addr.AddrPort()
			}
			if addr, ok := info.Conn.RemoteAddr().(*net.TCPAddr); ok {
				raddr = addr.AddrPort()
			}
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(traceCtx, trace))

	// Arrange for the inner context to be canceled
	// when the outer context is done.
	//
	// This must be after req.WithContext to avoid
	// a data race in the context itself.
	go func() {
		defer cancel()
		select {
		case <-req.Context().Done():
		case <-traceCtx.Done():
		}
	}()

	// Perform the request
	resp, err := client.Do(req)

	// Gather the local and remote endpoints while holding the mutex
	// to avoid data-racing with the tracing goroutine.
	mu.Lock()
	epnts := &Endpoints{LocalAddr: laddr, RemoteAddr: raddr}
	mu.Unlock()

	// Return the results to the caller.
	return resp, epnts, err
}
