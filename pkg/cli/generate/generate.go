//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// `rbmk generate` implementation.
//

// Package generate implements the `rbmk generate` Command.
package generate

import (
	_ "embed"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk generate` Command.
func NewCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands(
		"generate",
		markdown.LazyMaybeRender(readme),
		map[string]cliutils.Command{
			"stun_lookup": newSTUNLookupCommand(),
		},
	)
}
