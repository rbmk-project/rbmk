// SPDX-License-Identifier: GPL-3.0-or-later

package pipe

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

// newWriteCommand creates the `rbmk pipe write` command.
func newWriteCommand() cliutils.Command {
	return writeCommand{}
}

// writeCommand implements [cliutils.Command].
type writeCommand struct{}

var _ cliutils.Command = writeCommand{}

//go:embed write.md
var writeDocs string

// Help implements [cliutils.Command].
func (cmd writeCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(writeDocs))
	return nil
}

// Main implements [cliutils.Command].
func (cmd writeCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line arguments
	if len(argv) != 2 {
		err := errors.New("expected exactly one pipe name")
		fmt.Fprintf(env.Stderr(), "rbmk pipe write: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk pipe write --help' for usage.\n")
		return err
	}
	pipeName := argv[1]

	// 2. attempt connection with retries
	var (
		conn net.Conn
		err  error
	)
	delays := []time.Duration{5, 10, 20, 40, 80, 160, 320, 355} // ~1s total
	for _, delay := range delays {
		if conn, err = env.FS().DialUnix(pipeName); err == nil {
			break
		}
		time.Sleep(delay * time.Millisecond)
	}
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk pipe write: cannot connect to pipe: %s\n", err.Error())
		return err
	}
	defer conn.Close()

	// 3. copy data from stdin
	if _, err = io.Copy(conn, env.Stdin()); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk pipe write: cannot write data: %s\n", err.Error())
		return err
	}
	return nil
}
