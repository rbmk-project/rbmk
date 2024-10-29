// SPDX-License-Identifier: GPL-3.0-or-later

// Package dig implements the `rbmk dig` command.
package dig

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/rbmk-project/rbmk/cmd/internal/cliutils"
)

// NewCommand creates the `rbmk dig` [cliutils.Command].
func NewCommand() cliutils.Command {
	return command{}
}

// command implements [cliutils.command].
type command struct{}

var _ cliutils.Command = command{}

//go:embed README.txt
var readme string

// Help implements [cliutils.Command].
func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

// Main implements [cliutils.Command].
func (cmd command) Main(ctx context.Context, argv ...string) error {
	// honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// for now, the command is not implemented
	return errors.New("not implemented")
}
