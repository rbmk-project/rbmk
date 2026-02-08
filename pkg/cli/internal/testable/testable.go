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
	"io"
	"os"
	"sync"

	"github.com/rbmk-project/rbmk/internal/testablenet"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/rbmk-project/rbmk/pkg/common/fsx"
)

// DialContextFunc is an alias for [testablenet.DialContextFunc].
type DialContextFunc = testablenet.DialContextFunc

// DialContextProvider is an alias for [testablenet.DialContextProvider].
type DialContextProvider = testablenet.DialContextProvider

// DialContext is the singleton allowing to override the function used
// to establish network connections without data races.
//
// By default, we use the standard library to dial connections.
var DialContext = testablenet.DialContext

// RootCAsProvider is an alias for [testablenet.RootCAsProvider].
type RootCAsProvider = testablenet.RootCAsProvider

// RootCAs is the singleton allowing to override the root CAs.
//
// By default, we use the system root CAs.
var RootCAs = testablenet.RootCAs

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
