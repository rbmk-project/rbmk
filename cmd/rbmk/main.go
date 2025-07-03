// SPDX-License-Identifier: GPL-3.0-or-later

// Command rbmk implements the `rbmk` command.
package main

import (
	_ "embed"
	"os"

	"github.com/rbmk-project/rbmk/pkg/cli"
	"github.com/rbmk-project/rbmk/pkg/common/climain"
)

var mainArgs = os.Args

func main() {
	climain.Run(cli.NewCommand(), os.Exit, mainArgs...)
}
