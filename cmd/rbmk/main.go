// SPDX-License-Identifier: GPL-3.0-or-later

// Command rbmk implements the `rbmk` command.
package main

import (
	_ "embed"
	"os"

	"github.com/rbmk-project/common/climain"
	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/cli/dig"
)

var mainArgs = os.Args

func main() {
	climain.Run(newCommand(), os.Exit, mainArgs...)
}

//go:embed README.txt
var readme string

// newCommand constructs a new [cliutils.Command] for the `rbmk` command.
func newCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands("rbmk", readme, map[string]cliutils.Command{
		"dig": dig.NewCommand(),
	})
}
