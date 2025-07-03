// SPDX-License-Identifier: GPL-3.0-or-later

// Package intro implements the `rbmk intro` Command.
package intro

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk intro` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

// Help implements [cliutils.Command].
func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	return cmd.Main(context.Background(), env, argv...)
}

// Main implements [cliutils.Command].
func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}
