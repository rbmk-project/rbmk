// SPDX-License-Identifier: GPL-3.0-or-later

// Package stun implements the `rbmk stun` command.
package stun

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rbmk-project/common/cliutils"
	"github.com/spf13/pflag"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk stun` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

// Help implements [cliutils.Command].
func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", readme)
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

	// 7. process the endpoint argument
	task.Endpoint = args[0]

	// 8. handle --logs flag
	switch *logfile {
	case "":
		// nothing
	case "-":
		task.LogsWriter = env.Stdout()
	default:
		filep, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			err = fmt.Errorf("cannot open log file: %w", err)
			fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
			return err
		}
		defer filep.Close()
		task.LogsWriter = io.MultiWriter(task.LogsWriter, filep)
	}

	// 9. run the task
	if err := task.Run(ctx); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk stun: %s\n", err.Error())
		return err
	}
	return nil
}
