// SPDX-License-Identifier: GPL-3.0-or-later

// Package mkdir implements the `rbmk mkdir` command.
package mkdir

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/rbmk-project/common/cliutils"
	"github.com/spf13/pflag"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk mkdir` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

func (cmd command) Main(ctx context.Context, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk mkdir", pflag.ContinueOnError)
	parents := clip.BoolP("parents", "p", false, "create parent directories as needed")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk mkdir: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk mkdir --help` for usage.\n")
		return err
	}

	// 3. ensure we have at least one directory to create
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("missing directory name")
		fmt.Fprintf(os.Stderr, "rbmk mkdir: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk mkdir --help` for usage.\n")
		return err
	}

	// 4. create each directory
	for _, dir := range args {
		mkdirfn := os.Mkdir
		if *parents {
			mkdirfn = os.MkdirAll
		}
		if err := mkdirfn(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "rbmk mkdir: %s\n", err.Error())
			return err
		}
	}

	return nil
}
