// SPDX-License-Identifier: GPL-3.0-or-later

// Package cliutils provides utilities for building command-line interfaces.
package cliutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rbmk-project/rbmk/pkg/common/fsx"
)

// Environment is the environment for executing a [Command].
type Environment interface {
	// FS returns the virtual filesystem to use.
	FS() fsx.FS

	// Stdin returns the stdin reader to use.
	Stdin() io.Reader

	// Stderr returns the stderr writer to use.
	Stderr() io.Writer

	// Stdout returns the stdout writer to use.
	Stdout() io.Writer
}

// StandardEnvironment is the standard implementation of [Environment].
type StandardEnvironment struct{}

// Ensure that [StandardEnvironment] implements [Environment].
var _ Environment = StandardEnvironment{}

// FS implements [Environment].
func (se StandardEnvironment) FS() fsx.FS {
	return fsx.OsFS{}
}

// Stdin implements [Environment].
func (se StandardEnvironment) Stdin() io.Reader {
	return os.Stdin
}

// Stderr implements [Environment].
func (se StandardEnvironment) Stderr() io.Writer {
	return os.Stderr
}

// Stdout implements [Environment].
func (se StandardEnvironment) Stdout() io.Writer {
	return os.Stdout
}

// Command is an rbmk command-line command.
type Command interface {
	// Help prints the help for the command on the stdout.
	Help(env Environment, argv ...string) error

	// Main executes the command main function.
	Main(ctx context.Context, env Environment, argv ...string) error
}

// CommandWithSubCommands is a [Command] that contains subcommands.
//
// It works as follows:
//
// 1. It automatically handles invocation with no arguments by printing help.
//
// 2. It handles `-h` and `--help` by printing help.
//
// 3. It handles `help [COMMAND...]` by printing help either for the
// command itself or for the selected subcommmand.
//
// 4. It handles `COMMAND...` by redirecting execution to the subcommand.
//
// Construct using [NewCommandWithSubCommands].
type CommandWithSubCommands struct {
	// commands is the map of subcommands.
	commands map[string]Command

	// name is the full name of this command.
	name string

	// renderer is the help renderer.
	renderer LazyHelpRenderer
}

// LazyHelpRenderer renders the help possibly adding colours and formatting.
type LazyHelpRenderer interface {
	Help() string
}

// LazyHelpRendererFunc is a function that implements [LazyHelpRenderer].
type LazyHelpRendererFunc func() string

// Ensure that [LazyHelpRendererFunc] implements [LazyHelpRenderer].
var _ LazyHelpRenderer = LazyHelpRendererFunc(nil)

// Help implements HelpRenderer.
func (fx LazyHelpRendererFunc) Help() string {
	return fx()
}

// NewCommandWithSubCommands constructs a [CommandWithSubCommands].
//
// The name argument contains the full name of this command (e.g., `rbmk run`).
//
// The renderer argument renders the help on demand.
//
// The commands argument contains the implemented subcommands.
func NewCommandWithSubCommands(name string,
	renderer LazyHelpRenderer, commands map[string]Command) CommandWithSubCommands {
	return CommandWithSubCommands{
		commands: commands,
		name:     name,
		renderer: renderer,
	}
}

var _ Command = CommandWithSubCommands{}

// Help implements [Command].
func (c CommandWithSubCommands) Help(env Environment, argv ...string) error {
	// 1. case where we're invoked with no arguments
	if len(argv) < 2 {
		fmt.Fprintf(env.Stderr(), "%s\n", c.renderer.Help())
		return nil
	}

	// 2. obtain the command to print help for
	command := c.getCommand(argv[1])

	// 3. print the command help
	return command.Help(env, argv[1:]...)
}

// Main implements [Command].
func (c CommandWithSubCommands) Main(ctx context.Context, env Environment, argv ...string) error {
	switch {
	case len(argv) < 2:
		return c.Help(env)

	case argv[1] == "--help":
		return c.Help(env)
	case argv[1] == "-h":
		return c.Help(env)
	case argv[1] == "help":
		return c.Help(env, argv[1:]...)

	default:
		command := c.getCommand(argv[1])
		return command.Main(ctx, env, argv[1:]...)
	}
}

// getCommand returns the [Command] for the given name. If no command exists, we
// return a default [Command] that prints an error and gives usage hints.
func (c CommandWithSubCommands) getCommand(name string) Command {
	command := c.commands[name]
	if command == nil {
		command = newDefaultCommand(c.name)
	}
	return command
}

// defaultCommand is the default [Command] returned by [CommandWithSubCommands]
// when the argv[0] value does not identify any valid subcommand.
type defaultCommand struct {
	name string
}

// newDefaultCommand constructs a new [defaultCommand] instance.
func newDefaultCommand(name string) defaultCommand {
	return defaultCommand{name}
}

var _ Command = defaultCommand{}

// Help implements [Command].
func (dc defaultCommand) Help(env Environment, argv ...string) error {
	err := errors.New("no such help topic")
	fmt.Fprintf(env.Stderr(), "%s help: %s.\nTry `%s --help`.\n", dc.name, err.Error(), dc.name)
	return err
}

// Main implements [Command].
func (dc defaultCommand) Main(ctx context.Context, env Environment, argv ...string) error {
	err := errors.New("no such command")
	fmt.Fprintf(env.Stderr(), "%s %s: %s.\nTry `%s --help`.\n", dc.name, argv[0], err.Error(), dc.name)
	return err
}

// HelpRequested reads the argv and returns whether it contains
// one of `-h`, `--help`, in any position, or `help` as the first
// element in the vector. If this happens a subcommand should
// invoke its own help method to print help.
func HelpRequested(argv ...string) bool {
	for _, arg := range argv {
		switch {
		case arg == "-h" || arg == "--help":
			return true
		}
	}
	return len(argv) > 1 && argv[1] == "help"
}
