// SPDX-License-Identifier: GPL-3.0-or-later

// Package plugin implements the `rbmk plugin` command.
package plugin

import "github.com/rbmk-project/common/cliutils"

// NewCommand creates the `rbmk plugin` Command.
func NewCommand() cliutils.Command {
	return newCommand()
}
