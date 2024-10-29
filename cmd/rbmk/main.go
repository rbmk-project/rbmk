// SPDX-License-Identifier: GPL-3.0-or-later

// Command rbmk implements the `rbmk` command.
package main

import (
	_ "embed"
	"os"

	"github.com/rbmk-project/rbmk/cmd/internal/climain"
	"github.com/rbmk-project/rbmk/cmd/internal/cliutils"
	"github.com/rbmk-project/rbmk/cmd/rbmk/internal/dig"
)

var mainArgs = os.Args

func main() {
	climain.Run(newCommand(), os.Exit, mainArgs...)
}

//go:embed README.txt
var readme string

// newCommand constructs a new [cliutils.Command] for the `buresu` command.
func newCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands("buresu", readme, map[string]cliutils.Command{
		"dig": dig.NewCommand(),
	})
}
