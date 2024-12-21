// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package rootcmd contains shared code to
implement the `rbmk` command.

This code is shared between the following packages:

1. [github.com/rbmk-project/rbmk/pkg/cli/sh].

2. [github.com/rbmk-project/rbmk/pkg/cli].

The former package implements the `rbmk sh` command,
while the latter implements the `rbmk` command.

Both packages need to have access to the list of internal
commands as well as to the text printed on `--help`.
*/
package rootcmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/kballard/go-shellquote"
	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/pkg/cli/cat"
	"github.com/rbmk-project/rbmk/pkg/cli/curl"
	"github.com/rbmk-project/rbmk/pkg/cli/dig"
	"github.com/rbmk-project/rbmk/pkg/cli/generate"
	"github.com/rbmk-project/rbmk/pkg/cli/head"
	"github.com/rbmk-project/rbmk/pkg/cli/intro"
	"github.com/rbmk-project/rbmk/pkg/cli/ipuniq"
	"github.com/rbmk-project/rbmk/pkg/cli/mkdir"
	"github.com/rbmk-project/rbmk/pkg/cli/mv"
	"github.com/rbmk-project/rbmk/pkg/cli/nc"
	"github.com/rbmk-project/rbmk/pkg/cli/pipe"
	"github.com/rbmk-project/rbmk/pkg/cli/rm"
	"github.com/rbmk-project/rbmk/pkg/cli/stun"
	"github.com/rbmk-project/rbmk/pkg/cli/tar"
	"github.com/rbmk-project/rbmk/pkg/cli/timestamp"
	"github.com/rbmk-project/rbmk/pkg/cli/tutorial"
	"github.com/rbmk-project/rbmk/pkg/cli/version"
)

//go:embed README.md
var readme string

// HelpText returns the text to be printed when the `--help`
// flag is passed to the `rbmk` command. The text may be rendered
// using the markdown package, depending on the build tags.
func HelpText() string {
	return readme
}

// CommandsWithoutSh returns a new map containing the mapping
// between the name of a command and the command itself.
//
// The returned map DOES NOT include the `sh` command, which
// is a special case, which the packages using this func could
// choose to implement differently (and how they choose to
// implement it is not this function's concern anyway).
func CommandsWithoutSh() map[string]cliutils.Command {
	return map[string]cliutils.Command{
		"cat":       &cmdwrapper{cat.NewCommand()},
		"curl":      &cmdwrapper{curl.NewCommand()},
		"dig":       &cmdwrapper{dig.NewCommand()},
		"generate":  &cmdwrapper{generate.NewCommand()},
		"head":      head.NewCommand(),
		"intro":     &cmdwrapper{intro.NewCommand()},
		"ipuniq":    &cmdwrapper{ipuniq.NewCommand()},
		"mkdir":     &cmdwrapper{mkdir.NewCommand()},
		"mv":        &cmdwrapper{mv.NewCommand()},
		"nc":        &cmdwrapper{nc.NewCommand()},
		"pipe":      &cmdwrapper{pipe.NewCommand()},
		"rm":        &cmdwrapper{rm.NewCommand()},
		"stun":      &cmdwrapper{stun.NewCommand()},
		"tar":       &cmdwrapper{tar.NewCommand()},
		"timestamp": &cmdwrapper{timestamp.NewCommand()},
		"tutorial":  &cmdwrapper{tutorial.NewCommand()},
		"version":   &cmdwrapper{version.NewCommand()},
	}
}

type cmdwrapper struct {
	cmd cliutils.Command
}

// Help implements [cliutils.Command].
func (cw cmdwrapper) Help(env cliutils.Environment, argv ...string) error {
	return cw.cmd.Help(env, argv...)
}

// Main implements [cliutils.Command].
func (cw cmdwrapper) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	if os.Getenv("RBMK_TRACE") == "1" {
		fmt.Fprintf(os.Stderr, "+ %s\n", shellquote.Join(argv...))
	}
	return cw.cmd.Main(ctx, env, argv...)
}
