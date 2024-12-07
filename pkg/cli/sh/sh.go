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
	"github.com/rbmk-project/rbmk/internal/markdown"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk sh` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.TryRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. Honour requests for printing the help.
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. Ensure we have exactly one script to run.
	if len(argv) != 2 {
		err := errors.New("expected exactly one script argument")
		fmt.Fprintf(env.Stderr(), "rbmk sh: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk sh --help` for usage.\n")
		return err
	}

	// 3. Open and parse the shell script.
	scriptPath := argv[1]
	filep, err := os.Open(scriptPath)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: cannot open script: %s\n", err.Error())
		return err
	}
	defer filep.Close()

	parser := syntax.NewParser()
	prog, err := parser.Parse(filep, scriptPath)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: cannot parse script: %s\n", err.Error())
		return err
	}

	// 4. Ensure the RBMK_EXE environment variable is set.
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("rbmk sh: cannot determine rbmk path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("rbmk sh: cannot determine absolute rbmk path: %w", err)
	}
	os.Setenv("RBMK_EXE", exePath)

	// 5. Create the shell interpreter.
	runner, err := interp.New(
		interp.StdIO(env.Stdin(), env.Stdout(), env.Stderr()),
		interp.Env(expand.FuncEnviron(os.Getenv)),
	)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: cannot create interpreter: %s\n", err.Error())
		return err
	}

	// 6. Finally, run the shell script.
	err = runner.Run(ctx, prog)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: %s\n", err.Error())
		return err
	}
	return nil
}
