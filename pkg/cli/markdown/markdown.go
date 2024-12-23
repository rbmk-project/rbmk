// Package markdown implements the `rbmk markdown` command.
package markdown

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk markdown` Command.
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

	// 2. ensure there are no command line arguments
	if len(argv) > 1 {
		err := fmt.Errorf("expected no positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk markdown: %s\n", err)
		fmt.Fprintf(env.Stderr(), "Run `rbmk markdown --help` for usage.\n")
		return err
	}

	// 3. read all content from stdin
	input, err := io.ReadAll(env.Stdin())
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk markdown: %s\n", err)
		return err
	}

	// 4. render and write to stdout
	fmt.Fprintf(env.Stdout(), "%s", markdown.MaybeRender(string(input)))
	return nil
}
