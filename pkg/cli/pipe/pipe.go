// SPDX-License-Identifier: GPL-3.0-or-later

// Package pipe implements the `rbmk pipe` command.
package pipe

import (
	_ "embed"

	"github.com/rbmk-project/rbmk/pkg/cli/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
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
