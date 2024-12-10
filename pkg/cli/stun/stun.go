// SPDX-License-Identifier: GPL-3.0-or-later

// Package stun implements the `rbmk stun` command.
package stun

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/closepool"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk stun` Command.
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

	// 2. create initial task with defaults
	task := &Task{
		LogsWriter: io.Discard,
		Output:     env.Stdout(),
	}

	// 3. create command line parser
	clip := pflag.NewFlagSet("rbmk stun", pflag.ContinueOnError)

	// 4. add flags to the parser
	logfile := clip.String("logs", "", "path where to write structured logs")
	measure := clip.Bool("measure", false, "do not exit 1 on measurement failure")

	// 5. parse command line arguments
	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk stun --help` for usage.\n")
		return err
	}

	// 6. make sure we have exactly one endpoint argument
	args := clip.Args()
	if len(args) != 1 {
		err := errors.New("expected exactly one STUN endpoint to measure\n")
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk stun --help` for usage.\n")
		return err
	}

	// 7. finish filling up the task
	task.Endpoint = args[0]

	// 8. handle --logs flag
	var filepool closepool.Pool
	switch *logfile {
	case "":
		// nothing
	case "-":
		task.LogsWriter = env.Stdout()
	default:
		filep, err := env.FS().OpenFile(*logfile, fsx.O_CREATE|fsx.O_WRONLY|fsx.O_APPEND, 0600)
		if err != nil {
			err = fmt.Errorf("cannot open log file: %w", err)
			fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
			return err
		}
		filepool.Add(filep)
		task.LogsWriter = io.MultiWriter(task.LogsWriter, filep)
	}

	// 9. run the task and honour the `--measure` flag
	err := task.Run(ctx)
	if err != nil && *measure {
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "rbmk stun: not failing because you specified --measure\n")
		err = nil
	}

	// 10. ensure we close the opened files
	if err2 := filepool.Close(); err2 != nil {
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err2.Error())
		return err2
	}

	// 11. handle error when running the task
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
		return err
	}
	return nil
}
