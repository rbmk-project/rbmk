// SPDX-License-Identifier: GPL-3.0-or-later

// Package cat implements the `rbmk cat` command.
package cat

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rbmk-project/common/cliutils"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk cat` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", readme)
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. ensure we have at least one file to read
	if len(argv) < 2 {
		err := errors.New("expected one or more files to concatenate")
		fmt.Fprintf(env.Stderr(), "rbmk cat: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk cat --help` for usage.\n")
		return err
	}

	// 3. concatenate each file to stdout
	for _, path := range argv[1:] {
		if err := catFile(env, path); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk cat: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func catFile(env cliutils.Environment, path string) error {
	filep, err := os.Open(path)
	if err != nil {
		return err
	}
	defer filep.Close()
	_, err = io.Copy(env.Stdout(), filep)
	return err
}
