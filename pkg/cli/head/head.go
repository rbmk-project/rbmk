// SPDX-License-Identifier: GPL-3.0-or-later

// Package head implements the `rbmk head` command.
package head

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"

	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk head` Command.
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
	clip := pflag.NewFlagSet("rbmk head", pflag.ContinueOnError)
	lines := clip.UintP("lines", "n", 10, "number of lines to print")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk head: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk head --help` for usage.\n")
		return err
	}

	// 3. collect the files to read from, if any. Otherwise,
	// we will read from the standard input.
	args := clip.Args()
	if len(args) <= 0 {
		args = append(args, "-")
	}

	// 4. read from each file
	for _, fname := range args {
		if err := readHead(env, fname, *lines); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk head: %s\n", err.Error())
			return err
		}
	}
	return nil
}

// readHead reads up to count lines from the given file and prints them
// to stdout. The special filename "-" means read from stdin.
func readHead(env cliutils.Environment, fname string, count uint) error {
	var reader io.Reader
	if fname != "-" {
		filep, err := env.FS().Open(fname)
		if err != nil {
			return err
		}
		defer filep.Close()
		reader = filep
	} else {
		reader = env.Stdin()
	}

	scanner := bufio.NewScanner(reader)
	for linenum := uint(0); linenum < count && scanner.Scan(); linenum++ {
		fmt.Fprintln(env.Stdout(), scanner.Text())
	}
	return scanner.Err()
}
