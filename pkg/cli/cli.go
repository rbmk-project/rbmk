// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli implements the `rbmk` command.
package cli

import (
	_ "embed"

	"github.com/rbmk-project/rbmk/pkg/cli/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/cli/internal/rootcmd"
	"github.com/rbmk-project/rbmk/pkg/cli/sh"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

// NewCommand constructs a new [cliutils.Command] for the `rbmk` command.
func NewCommand() cliutils.Command {
	directory := rootcmd.CommandsWithoutSh()
	directory["sh"] = sh.NewCommand()
	return cliutils.NewCommandWithSubCommands(
		"rbmk", markdown.LazyMaybeRender(rootcmd.HelpText()), directory)
}
