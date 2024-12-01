// SPDX-License-Identifier: GPL-3.0-or-later

// Package timestamp implements the `rbmk timestamp` command.
package timestamp

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/rbmk-project/common/cliutils"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk timestamp` Command.
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

	// 2. ensure no extra arguments
	if len(argv) > 1 {
		fmt.Fprintf(os.Stderr, "rbmk timestamp: unexpected arguments\n")
		fmt.Fprintf(os.Stderr, "Run `rbmk timestamp --help` for usage.\n")
		return fmt.Errorf("unexpected arguments")
	}

	// 3. print UTC timestamp in compact format
	fmt.Fprintf(os.Stdout, "%s\n", time.Now().UTC().Format("20060102T150405Z"))
	return nil
}
