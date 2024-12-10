// SPDX-License-Identifier: GPL-3.0-or-later

// Package mkdir implements the `rbmk mkdir` command.
package mkdir

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk mkdir` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse the command line flags
	clip := pflag.NewFlagSet("rbmk mkdir", pflag.ContinueOnError)
	parents := clip.BoolP("parents", "p", false, "create parent directories as needed")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk mkdir: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk mkdir --help` for usage.\n")
		return err
	}

	// 3. ensure we have at least one directory to create
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("expected one or more directories to create")
		fmt.Fprintf(env.Stderr(), "rbmk mkdir: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk mkdir --help` for usage.\n")
		return err
	}

	// 4. create each directory
	for _, dir := range args {
		mkdirfn := env.FS().Mkdir
		if *parents {
			mkdirfn = env.FS().MkdirAll
		}
		if err := mkdirfn(dir, 0755); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk mkdir: %s\n", err.Error())
			return err
		}
	}
	return nil
}
