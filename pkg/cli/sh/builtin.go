// SPDX-License-Identifier: GPL-3.0-or-later

package sh

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/internal/rootcmd"
	"mvdan.cc/sh/v3/interp"
)

// builtInMiddleware is the middleware to execute built-in commands.
type builtInMiddleware func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc

// newBuiltInMiddleware creates a new built-in middleware for
// executing built-in commands with the shell.
func newBuiltInMiddleware() builtInMiddleware {
	return func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			// 1. ensure we have a command to run and that such a
			// command is indeed the "rbmk" internal command.
			if len(args) < 1 {
				return errors.New("no command to run")
			}
			if args[0] != "rbmk" {
				return fmt.Errorf("%s: command not found", args[0])
			}

			// 2. construct the subcommand environment.
			env := newBuiltInEnvironment(interp.HandlerCtx(ctx))

			// 3. construct the root command to switch depending on the
			// actual `rbmk` subcommand being invoked.
			directory := rootcmd.CommandsWithoutSh()
			directory["sh"] = builtInShCommand{}
			root := cliutils.NewCommandWithSubCommands(
				"rbmk", markdown.LazyMaybeRender(rootcmd.HelpText()), directory)

			// 4. execute the root command and return the result
			return root.Main(ctx, env, args...)
		}
	}
}

// builtInEnvironment contains the environment for executing built-in commands.
type builtInEnvironment struct {
	// fs is the file system to use.
	fs fsx.FS

	// stdin is the standard input.
	stdin io.Reader

	// stdout is the standard output.
	stdout io.Writer

	// stderr is the standard error.
	stderr io.Writer
}

// newBuiltInEnvironment creates a new built-in environment.
//
// Uses:
//
// 1. [fsx.NewChdirFS] to simulate chdir into the current directory.
//
// 2. the shells's current stdin, stdout, and stderr.
//
// We ignore the shell environment since we don't actually use it.
func newBuiltInEnvironment(shCtx interp.HandlerContext) *builtInEnvironment {
	return &builtInEnvironment{
		fs:     fsx.NewChdirFS(fsx.OsFS{}, shCtx.Dir),
		stdin:  shCtx.Stdin,
		stdout: shCtx.Stdout,
		stderr: shCtx.Stderr,
	}
}

// Ensure that builtInEnvironment implements [cliutils.Environment].
var _ cliutils.Environment = &builtInEnvironment{}

// FS implements [cliutils.Environment].
func (env *builtInEnvironment) FS() fsx.FS {
	return env.fs
}

// Stderr implements [cliutils.Environment].
func (env *builtInEnvironment) Stderr() io.Writer {
	return env.stderr
}

// Stdin implements [cliutils.Environment].
func (env *builtInEnvironment) Stdin() io.Reader {
	return env.stdin
}

// Stdout implements [cliutils.Environment].
func (env *builtInEnvironment) Stdout() io.Writer {
	return env.stdout
}

// builtInShCommand is the built-in `sh` command executed
// from inside the `rbmk sh` environment. We do not permit
// executing the shell inside the shell because that has
// not been tested, and it would probably not WAI.
type builtInShCommand struct{}

// Ensure that [builtInShCommand] implements [cliutils.Command].
var _ cliutils.Command = builtInShCommand{}

// Help implements [cliutils.Command].
func (cmd builtInShCommand) Help(env cliutils.Environment, argv ...string) error {
	return NewCommand().Help(env, argv...)
}

// Main implements [cliutils.Command].
func (cmd builtInShCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. Honour requests for printing the help.
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. otherwise prevent re-execution of the shell
	err := errors.New("cannot execute `rbmk sh` inside `rbmk sh`")
	fmt.Fprintf(env.Stderr(), "rbmk sh: %s\n", err.Error())
	return err
}
