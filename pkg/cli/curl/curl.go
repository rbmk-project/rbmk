// SPDX-License-Identifier: GPL-3.0-or-later

// Package curl implements the `rbmk curl` command.
package curl

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/closepool"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

// NewCommand creates the `rbmk curl` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

//go:embed README.md
var readme string

// Help implements cliutils.Command.
func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

// Main implements cliutils.Command.
func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. create initial task with defaults
	task := &Task{
		LogsWriter:    io.Discard,
		MaxTime:       30 * time.Second,
		Method:        "GET",
		Output:        env.Stdout(),
		ResolveMap:    make(map[string]string),
		URL:           "",
		VerboseOutput: io.Discard,
	}

	// 3. create command line parser
	clip := pflag.NewFlagSet("rbmk curl", pflag.ContinueOnError)

	// 4. add flags to the parser
	logfile := clip.String("logs", "", "path where to write structured logs")
	maxTime := clip.Int64("max-time", 30, "maximum time to wait for the operation to finish")
	measure := clip.Bool("measure", false, "do not exit 1 on measurement failure")
	output := clip.StringP("output", "o", "", "write to file instead of stdout")
	method := clip.StringP("request", "X", "GET", "HTTP request method")
	resolve := clip.StringArray("resolve", nil, "use addr instead of DNS")
	verbose := clip.BoolP("verbose", "v", false, "make more talkative")

	// 5. parse command line arguments
	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 6. make sure we have exactly one URL argument
	positional := clip.Args()
	if len(positional) != 1 {
		err := errors.New("expected exactly one URL argument")
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 7. process the URL argument
	task.URL = positional[0]
	if !strings.HasPrefix(task.URL, "http://") && !strings.HasPrefix(task.URL, "https://") {
		err := errors.New("URL scheme must be http:// or https://")
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk curl --help` for usage.\n")
		return err
	}

	// 8. process --resolve entries by splitting
	for _, entry := range *resolve {
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 {
			err := fmt.Errorf("invalid --resolve value: %s", entry)
			fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
			fmt.Fprintf(env.Stderr(), "Run `rbmk curl --help` for usage.\n")
			return err
		}
		// Implementation note: we ignore the port since our
		// LookupHost function does not know the port.
		task.ResolveMap[parts[0]] = parts[2]
	}

	// 9. process other flags
	task.MaxTime = time.Duration(*maxTime) * time.Second
	task.Method = *method
	if *verbose {
		task.VerboseOutput = env.Stderr()
	}

	// 10. handle --logs flag
	var filepool closepool.Pool
	switch *logfile {
	case "":
		// nothing
	case "-":
		task.LogsWriter = env.Stdout()
	default:
		filep, err := env.FS().OpenFile(*logfile, fsx.O_CREATE|fsx.O_WRONLY|fsx.O_APPEND, 0600)
		if err != nil {
			err = fmt.Errorf("cannot open log file: %w", err)
			fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
			return err
		}
		filepool.Add(filep)
		task.LogsWriter = io.MultiWriter(task.LogsWriter, filep)
	}

	// 11. handle -o/--output flag
	if *output != "" {
		filep, err := env.FS().OpenFile(*output, fsx.O_CREATE|fsx.O_WRONLY|fsx.O_TRUNC, 0600)
		if err != nil {
			err = fmt.Errorf("cannot create output file: %w", err)
			fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
			return err
		}
		filepool.Add(filep)
		task.Output = filep
	}

	// 12. run the task and honour the `--measure` flag
	err := task.Run(ctx)
	if err != nil && *measure {
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "rbmk curl: not failing because you specified --measure\n")
		err = nil
	}

	// 13. ensure we close the opened files
	if err2 := filepool.Close(); err2 != nil {
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err2.Error())
		return err2
	}

	// 14. handle error when running the task
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk curl: %s\n", err.Error())
		return err
	}
	return nil
}
