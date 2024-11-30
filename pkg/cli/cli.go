// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli implements the `rbmk` command.
package cli

import (
	_ "embed"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/pkg/cli/curl"
	"github.com/rbmk-project/rbmk/pkg/cli/dig"
	"github.com/rbmk-project/rbmk/pkg/cli/intro"
	"github.com/rbmk-project/rbmk/pkg/cli/tutorial"
)

//go:embed README.txt
var readme string

// NewCommand constructs a new [cliutils.Command] for the `rbmk` command.
func NewCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands("rbmk", readme, map[string]cliutils.Command{
		"curl":     curl.NewCommand(),
		"dig":      dig.NewCommand(),
		"intro":    intro.NewCommand(),
		"tutorial": tutorial.NewCommand(),
	})
}
