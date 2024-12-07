// SPDX-License-Identifier: GPL-3.0-or-later

package intro

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk intro` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	return cmd.Main(context.Background(), env, argv...)
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.TryRender(readme))
	return nil
}
