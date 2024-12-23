// SPDX-License-Identifier: GPL-3.0-or-later

// Package version implements the `rbmk version` command.
package version

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

// Version is the program version.
var Version string = "dev"

//go:embed README.md
var readme string

// NewCommand creates the `rbmk version` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

// Help implements [cliutils.Command].
func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

// Main implements [cliutils.Command].
func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. ensure there are no command line arguments
	if len(argv) > 1 {
		err := fmt.Errorf("expected no positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk version: %s\n", err)
		fmt.Fprintf(env.Stderr(), "Run `rbmk version --help` for usage.\n")
		return err
	}

	// 3. print the version
	fmt.Fprintln(env.Stdout(), Version)
	return nil
}
