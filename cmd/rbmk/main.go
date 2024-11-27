// SPDX-License-Identifier: GPL-3.0-or-later

// Command rbmk implements the `rbmk` command.
package main

import (
	_ "embed"
	"os"

	"github.com/rbmk-project/common/climain"
	"github.com/rbmk-project/rbmk/internal/cli"
)

var mainArgs = os.Args

func main() {
	climain.Run(cli.NewCommand(), os.Exit, mainArgs...)
}
