// SPDX-License-Identifier: GPL-3.0-or-later

package pipe

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
	"github.com/spf13/pflag"
)

// newReadCommand creates the `rbmk pipe read` command.
func newReadCommand() cliutils.Command {
	return readCommand{}
}

// readCommand implements [cliutils.Command].
type readCommand struct{}

var _ cliutils.Command = readCommand{}

//go:embed read.md
var readDocs string

// Help implements [cliutils.Command].
func (cmd readCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readDocs))
	return nil
}

// Main implements [cliutils.Command].
func (cmd readCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line
	clip := pflag.NewFlagSet("rbmk pipe read", pflag.ContinueOnError)
	writers := clip.Int("writers", 0, "number of writers to expect")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk pipe read: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk pipe read --help' for usage.\n")
		return err
	}

	// 3. validate arguments
	if *writers <= 0 {
		err := errors.New("you must specify a positive number of writers using --writers")
		fmt.Fprintf(env.Stderr(), "rbmk pipe read: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk pipe read --help' for usage.\n")
		return err
	}
	args := clip.Args()
	if len(args) != 1 {
		err := errors.New("expected exactly one pipe name")
		fmt.Fprintf(env.Stderr(), "rbmk pipe read: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk pipe read --help' for usage.\n")
		return err
	}
	pipeName := args[0]

	// 4. create and setup listener
	listener, err := env.FS().ListenUnix(pipeName)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk pipe read: cannot create pipe: %s\n", err.Error())
		return err
	}
	defer func() {
		listener.Close()
		env.FS().Remove(pipeName)
	}()

	// 5. protect stdout writes and handle completion
	var (
		stdoutMu sync.Mutex
		wg       sync.WaitGroup
	)
	wg.Add(*writers) // expect exact number of writers

	// 6. accept exactly N writers
	for count := 0; count < *writers; count++ {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk pipe read: accept failed: %s\n", err.Error())
			return err
		}

		go func() {
			defer wg.Done()
			defer conn.Close()

			// read from the connection line by line and
			// write to stdout in a goroutine safe way
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Text()
				stdoutMu.Lock()
				fmt.Fprintln(env.Stdout(), line)
				stdoutMu.Unlock()
			}

			if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
				fmt.Fprintf(env.Stderr(), "rbmk pipe read: read error: %s\n", err.Error())
			}
		}()
	}

	// 7. wait for all writers to complete
	wg.Wait()
	return nil
}
