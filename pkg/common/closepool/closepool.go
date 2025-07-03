// SPDX-License-Identifier: GPL-3.0-or-later

// Package closepool allows pooling [io.Closer] instances
// and closing them in a single operation.
package closepool

import (
	"errors"
	"io"
	"slices"
	"sync"
)

// CloserFunc allows to turn a suitable function into an [io.Closer].
type CloserFunc func() error

// Ensure that [CloserFunc] implements [io.Closer].
var _ io.Closer = CloserFunc(nil)

// Close implements io.Closer.
func (fx CloserFunc) Close() error {
	return fx()
}

// Pool allows pooling a set of [io.Closer].
//
// The zero value is ready to use.
type Pool struct {
	// handles contains the [io.Closer] to close.
	handles []io.Closer

	// mu provides mutual exclusion.
	mu sync.Mutex
}

// Add adds a given [io.Closer] to the pool.
func (p *Pool) Add(conn io.Closer) {
	p.mu.Lock()
	p.handles = append(p.handles, conn)
	p.mu.Unlock()
}

// Close closes all the [io.Closer] inside the pool iterating
// in backward order. Therefore, if one registers a TCP connection
// and then the corresponding TLS connection, the TLS connection
// is closed first. The returned error is the join of all the
// errors that occurred when closing connections.
func (p *Pool) Close() error {
	// Lock and copy the [io.Closer] to close.
	p.mu.Lock()
	handles := p.handles
	p.handles = nil
	p.mu.Unlock()

	// Close all the [io.Closer].
	var errv []error
	for _, handle := range slices.Backward(handles) {
		if err := handle.Close(); err != nil {
			errv = append(errv, err)
		}
	}
	return errors.Join(errv...)
}
