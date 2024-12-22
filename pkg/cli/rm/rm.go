// SPDX-License-Identifier: GPL-3.0-or-later

// Package rm implements the `rbmk rm` command.
package rm

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk rm` Command.
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

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk rm", pflag.ContinueOnError)
	recursive := clip.BoolP("recursive", "r", false, "remove directories and their contents recursively")
	force := clip.BoolP("force", "f", false, "ignore nonexistent-file errors")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk rm: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk rm --help` for usage.\n")
		return err
	}

	// 3. ensure we have at least one path to remove
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("expected one or more paths to remove")
		fmt.Fprintf(env.Stderr(), "rbmk rm: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk rm --help` for usage.\n")
		return err
	}

	// 4. remove each path
	for _, path := range args {
		if err := removePath(env, path, *recursive, *force); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk rm: %s\n", err.Error())
			return err
		}
	}
	return nil
}

// removePath removes a file or directory at the given path.
func removePath(env cliutils.Environment, path string, recursive bool, force bool) error {
	info, err := env.FS().Lstat(path)
	switch {
	case err != nil && fsx.IsNotExist(err) && force:
		return nil
	case err != nil:
		return err
	case info.IsDir() && !recursive:
		return fmt.Errorf("cannot remove %s: is a directory", path)
	case info.IsDir():
		return env.FS().RemoveAll(path)
	default:
		return env.FS().Remove(path)
	}
}
