// SPDX-License-Identifier: GPL-3.0-or-later

// Package mv implements the `rbmk mv` command.
package mv

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk mv` Command.
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

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk mv", pflag.ContinueOnError)

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk mv: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk mv --help` for usage.\n")
		return err
	}

	// 3. ensure we have at least two arguments
	args := clip.Args()
	if len(args) < 2 {
		err := errors.New("missing source and/or destination")
		fmt.Fprintf(env.Stderr(), "rbmk mv: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk mv --help` for usage.\n")
		return err
	}

	// 4. get sources and destination
	sources := args[:len(args)-1]
	dest := args[len(args)-1]

	// 5. handle multiple sources case
	if len(sources) > 1 {
		// Check if destination is a directory
		finfo, err := os.Stat(dest)
		if err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk mv: %s\n", err.Error())
			return err
		}
		if !finfo.IsDir() {
			err := errors.New("destination must be a directory when moving multiple sources")
			fmt.Fprintf(env.Stderr(), "rbmk mv: %s\n", err.Error())
			return err
		}
	}

	// 6. process each source
	for _, src := range sources {
		// ensure we move inside directory if last element is a directory
		target := dest
		if findo, err := os.Stat(dest); err == nil && findo.IsDir() {
			target = filepath.Join(dest, filepath.Base(src))
		}

		// actually rename the file
		if err := os.Rename(src, target); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk mv: %s\n", err.Error())
			return err
		}
	}
	return nil
}
