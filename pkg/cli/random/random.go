// SPDX-License-Identifier: GPL-3.0-or-later

// Package random implements the `rbmk random` command.
package random

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"

	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk random` Command.
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

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk random", pflag.ContinueOnError)
	nbytes := clip.Uint("bytes", 4, "number of random bytes to generate")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk random: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk random --help` for usage.\n")
		return err
	}

	// 3. ensure we have no positional arguments
	args := clip.Args()
	if len(args) != 0 {
		err := errors.New("expected zero positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk random: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk random --help` for usage.\n")
		return err
	}

	// 4. generate the random bytes
	buf := make([]byte, *nbytes)
	if _, err := rand.Read(buf); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk random: %s\n", err.Error())
		return err
	}

	// 5. output as hex
	fmt.Fprintf(env.Stdout(), "%x\n", buf)
	return nil
}
