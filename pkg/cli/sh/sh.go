// SPDX-License-Identifier: GPL-3.0-or-later

// Package sh implements the `rbmk sh` command.
package sh

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rbmk-project/common/cliutils"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk sh` Command.
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

	// 2. ensure we have exactly one script to run
	if len(argv) != 2 {
		err := errors.New("expected exactly one script argument")
		fmt.Fprintf(os.Stderr, "rbmk sh: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk sh --help` for usage.\n")
		return err
	}

	// 3. open and parse the script
	scriptPath := argv[1]
	f, err := os.Open(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbmk sh: cannot open script: %s\n", err.Error())
		return err
	}
	defer f.Close()

	parser := syntax.NewParser()
	prog, err := parser.Parse(f, scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbmk sh: cannot parse script: %s\n", err.Error())
		return err
	}

	// Ensure the RBMK_EXE environment variable is set
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine rbmk path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("cannot determine absolute rbmk path: %w", err)
	}
	os.Setenv("RBMK_EXE", exePath)

	// 5. create shell runner
	runner, err := interp.New(
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
		interp.Env(expand.FuncEnviron(os.Getenv)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbmk sh: cannot create interpreter: %s\n", err.Error())
		return err
	}

	// 6. run the script
	err = runner.Run(ctx, prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbmk sh: %s\n", err.Error())
		return err
	}

	return nil
}
