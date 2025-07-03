// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package testable provides thread-safe singletons for overriding
fundamental RBMK dependencies in integration tests.

The zero value of each singleton is ready to use and typically
uses the standard library. Overriding to a different value allows
to either use mocks or replacements such as the ones implemented
by the [github.com/rbmk-project/rbmk/pkg/x/netsim] package.
*/
package testable

import (
	"context"
	"crypto/x509"
	"io"
	"net"
	"os"
	"sync"

	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/rbmk-project/rbmk/pkg/common/fsx"
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
	fx := dcp.fx
	if fx == nil {
		fx = (&net.Dialer{}).DialContext
	}
	return fx
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

// Environment implements a testable [cliutils.Environment].
//
// The zero value is not ready to use; construct using [NewEnvironment].
type Environment struct {
	// fsp is the file system provider.
	fsp fsx.FS

	// mu protects stderr and stdout.
	mu sync.Mutex

	// stdin is the standard input stream.
	stdin io.Reader

	// stderr is the standard error stream.
	stderr io.Writer

	// stdout is the standard output stream.
	stdout io.Writer
}

// NewEnvironment creates a new [*Environment] instance.
func NewEnvironment() *Environment {
	return &Environment{
		fsp:    fsx.OsFS{},
		mu:     sync.Mutex{},
		stdin:  os.Stdin,
		stderr: os.Stderr,
		stdout: os.Stdout,
	}
}

// SetFS sets the file system provider.
func (env *Environment) SetFS(fsp fsx.FS) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.fsp = fsp
}

// SetStdin sets the standard input stream.
func (env *Environment) SetStdin(r io.Reader) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.stdin = r
}

// SetStderr sets the standard error stream.
func (env *Environment) SetStderr(w io.Writer) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.stderr = w
}

// SetStdout sets the standard output stream.
func (env *Environment) SetStdout(w io.Writer) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.stdout = w
}

// Ensure that [*Environment] implements [cliutils.Environment].
var _ cliutils.Environment = (*Environment)(nil)

// FS implements [cliutils.Environment].
func (env *Environment) FS() fsx.FS {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.fsp
}

// Stdin implements [cliutils.Environment].
func (env *Environment) Stdin() io.Reader {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.stdin
}

// Stderr implements [cliutils.Environment].
func (env *Environment) Stderr() io.Writer {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.stderr
}

// Stdout implements [cliutils.Environment].
func (env *Environment) Stdout() io.Writer {
	env.mu.Lock()
	defer env.mu.Unlock()
	return env.stdout
}
