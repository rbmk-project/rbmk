// SPDX-License-Identifier: GPL-3.0-or-later

// Package pipe implements the `rbmk pipe` command.
package pipe

import (
	_ "embed"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk pipe` Command.
func NewCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands(
		"pipe", markdown.LazyMaybeRender(readme),
		map[string]cliutils.Command{
			"read":  newReadCommand(),
			"write": newWriteCommand(),
		})
}
