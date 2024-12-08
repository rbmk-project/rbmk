// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli implements the `rbmk` command.
package cli

import (
	_ "embed"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/cli/cat"
	"github.com/rbmk-project/rbmk/pkg/cli/curl"
	"github.com/rbmk-project/rbmk/pkg/cli/dig"
	"github.com/rbmk-project/rbmk/pkg/cli/intro"
	"github.com/rbmk-project/rbmk/pkg/cli/ipuniq"
	"github.com/rbmk-project/rbmk/pkg/cli/mkdir"
	"github.com/rbmk-project/rbmk/pkg/cli/mv"
	"github.com/rbmk-project/rbmk/pkg/cli/pipe"
	"github.com/rbmk-project/rbmk/pkg/cli/rm"
	"github.com/rbmk-project/rbmk/pkg/cli/sh"
	"github.com/rbmk-project/rbmk/pkg/cli/stun"
	"github.com/rbmk-project/rbmk/pkg/cli/tar"
	"github.com/rbmk-project/rbmk/pkg/cli/timestamp"
	"github.com/rbmk-project/rbmk/pkg/cli/tutorial"
	"github.com/rbmk-project/rbmk/pkg/cli/version"
)

//go:embed README.md
var readme string

// NewCommand constructs a new [cliutils.Command] for the `rbmk` command.
func NewCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands(
		"rbmk", markdown.LazyMaybeRender(readme),
		map[string]cliutils.Command{
			"cat":       cat.NewCommand(),
			"curl":      curl.NewCommand(),
			"dig":       dig.NewCommand(),
			"intro":     intro.NewCommand(),
			"ipuniq":    ipuniq.NewCommand(),
			"mkdir":     mkdir.NewCommand(),
			"mv":        mv.NewCommand(),
			"pipe":      pipe.NewCommand(),
			"rm":        rm.NewCommand(),
			"sh":        sh.NewCommand(),
			"stun":      stun.NewCommand(),
			"tar":       tar.NewCommand(),
			"timestamp": timestamp.NewCommand(),
			"tutorial":  tutorial.NewCommand(),
			"version":   version.NewCommand(),
		})
}
