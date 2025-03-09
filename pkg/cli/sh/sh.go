// SPDX-License-Identifier: GPL-3.0-or-later

// Package sh implements the `rbmk sh` command.
package sh

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

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
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. Ensure we have exactly one script to run.
	if len(argv) < 2 {
		err := errors.New("expected a script with optional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk sh: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk sh --help` for usage.\n")
		return err
	}

	// 2. If the script is named `-h` or `--help` print help.
	if argv[1] == "-h" || argv[1] == "--help" {
		return cmd.Help(env, argv...)
	}

	// 3. Open and parse the shell script.
	scriptPath := argv[1]
	filep, err := env.FS().Open(scriptPath)
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

	// 4. Create the shell interpreter ensuring we properly use `--` to
	// ensure options get passed to the script itself.
	scriptParams := append([]string{"--"}, argv[2:]...)
	runner, err := interp.New(
		interp.StdIO(env.Stdin(), env.Stdout(), env.Stderr()),
		interp.Env(expand.FuncEnviron(osGetenvWrapper)),
		interp.ExecHandlers(newBuiltInMiddleware()),
		interp.Params(scriptParams...),
	)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: cannot create interpreter: %s\n", err.Error())
		return err
	}

	// 5. Finally, run the shell script.
	err = runner.Run(ctx, prog)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk sh: %s\n", err.Error())
		return err
	}
	return nil
}

// rbmkExeVarName is the name of the RBMK_EXE environment variable,
// which is shared by this package and its unit tests.
const rbmkExeVarName = "RBMK_EXE"

// osGetenvWrapper is a wrapper for [os.Getenv] that ensures that the
// RBMK_EXE environment variable is set to support scripts written
// before the release of RBMK v0.7.0.
//
// We use a wrapper rather than setting the environment explicitly
// because [os.Setenv] is NOT THREAD SAFE in Go. Thus, it seems more
// robust to avoid using it and potentially racing between it and
// [os.Getenv], used by the shell interpreter.
//
// See: https://www.edgedb.com/blog/c-stdlib-isn-t-threadsafe-and-even-safe-rust-didn-t-save-us
func osGetenvWrapper(key string) string {
	value := os.Getenv(key)
	if value == "" && key == rbmkExeVarName {
		value = "rbmk"
	}
	return value
}
