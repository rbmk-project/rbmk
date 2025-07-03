// SPDX-License-Identifier: GPL-3.0-or-later

// Package cat implements the `rbmk cat` command.
package cat

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk cat` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

// Help implements [cliutils.Command].
func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

// Main implements [cliutils.Command].
func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. if there's no file to read, use the stdin
	if len(argv) < 2 {
		argv = append(argv, "-")
	}

	// 3. concatenate each file to stdout
	for _, path := range argv[1:] {
		if err := catFile(env, path); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk cat: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func catFile(env cliutils.Environment, path string) error {
	var reader io.Reader
	if path != "-" {
		filep, err := env.FS().Open(path)
		if err != nil {
			return err
		}
		defer filep.Close()
		reader = filep
	} else {
		reader = env.Stdin()
	}
	_, err := io.Copy(env.Stdout(), reader)
	return err
}
