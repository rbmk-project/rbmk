// SPDX-License-Identifier: GPL-3.0-or-later

// Package timestamp implements the `rbmk timestamp` command.
package timestamp

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/rbmk-project/rbmk/pkg/cli/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk timestamp` Command.
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

	// 2. parse command line
	clip := pflag.NewFlagSet("rbmk timestamp", pflag.ContinueOnError)
	ffull := clip.Bool("full", false, "whether to emit timestamp with nanosecond precision")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk timestamp: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk timestamp --help' for usage.\n")
		return err
	}

	// 3. ensure no extra arguments
	if len(clip.Args()) > 1 {
		err := errors.New("expected no positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk timestamp: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk timestamp --help` for usage.\n")
		return err
	}

	// 4. select the timestamp format
	format := "20060102T150405Z"
	if *ffull {
		format = "20060102T150405.999999999Z"
	}

	// 5. print ISO8601 UTC timestamp in compact format
	fmt.Fprintf(env.Stdout(), "%s\n", time.Now().UTC().Format(format))
	return nil
}
