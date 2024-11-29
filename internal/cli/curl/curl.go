// SPDX-License-Identifier: GPL-3.0-or-later

// Package curl implements the `rbmk curl` command.
package curl

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/spf13/pflag"
)

// NewCommand creates the `rbmk curl` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

//go:embed README.txt
var readme string

// Help implements cliutils.Command.
func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

// Main implements cliutils.Command.
func (cmd command) Main(ctx context.Context, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// Implementation note: we care about testing whether we
	// produce the correct logs in several simulated conditions,
	// therefore, the `stdout` used by logs is overridable
	// through the `testable` package.
	//
	// On the contrary, we care much less about testing logging
	// the request and response, or other error and output messages,
	// so we just use `os.Stdout` and `os.Stderr` directly.
	testableStdout := testable.Stdout.Get()

	// 2. create initial task with defaults
	task := &Task{
		LogsWriter:    io.Discard,
		Method:        "GET",
		Output:        os.Stdout,
		ResolveMap:    make(map[string]string),
		URL:           "",
		VerboseOutput: io.Discard,
	}

	// 3. create command line parser
	clip := pflag.NewFlagSet("rbmk curl", pflag.ContinueOnError)

	// 4. add flags to the parser
	logfile := clip.String("logs", "", "path where to write structured logs")
	output := clip.StringP("output", "o", "", "write to file instead of stdout")
	method := clip.StringP("request", "X", "GET", "HTTP request method")
	resolve := clip.StringArray("resolve", nil, "use addr instead of DNS")
	verbose := clip.BoolP("verbose", "v", false, "make more talkative")

	// 5. parse command line arguments
	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 6. make sure we have exactly one URL argument
	positional := clip.Args()
	if len(positional) != 1 {
		err := errors.New("expected exactly one URL argument")
		fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 7. process the URL argument
	task.URL = positional[0]
	if !strings.HasPrefix(task.URL, "http://") && !strings.HasPrefix(task.URL, "https://") {
		err := errors.New("URL scheme must be http:// or https://")
		fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 8. process --resolve entries by splitting
	for _, entry := range *resolve {
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 {
			err := fmt.Errorf("invalid --resolve value: %s", entry)
			fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "Run `rbmk curl --help` for usage.\n")
			return err
		}
		// Implementation note: we ignore the port since our
		// LookupHost function does not know the port.
		task.ResolveMap[parts[0]] = parts[2]
	}

	// 9. process other flags
	task.Method = *method
	if *verbose {
		task.VerboseOutput = os.Stderr
	}

	// 10. handle --logs flag
	switch *logfile {
	case "":
		// nothing
	case "-":
		task.LogsWriter = testableStdout
	default:
		filep, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			err = fmt.Errorf("cannot open log file: %w", err)
			fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
			return err
		}
		defer filep.Close()
		task.LogsWriter = io.MultiWriter(task.LogsWriter, filep)
	}

	// 11. handle -o/--output flag
	if *output != "" {
		filep, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			err = fmt.Errorf("cannot create output file: %w", err)
			fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
			return err
		}
		defer filep.Close()
		task.Output = filep
	}

	// 12. run the task
	if err := task.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk curl: %s\n", err.Error())
		return err
	}

	return nil
}
