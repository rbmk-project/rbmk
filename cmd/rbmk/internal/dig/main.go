// SPDX-License-Identifier: GPL-3.0-or-later

// Package dig implements the `rbmk dig` command.
package dig

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rbmk-project/rbmk/cmd/internal/cliutils"
	"github.com/spf13/pflag"
)

// NewCommand creates the `rbmk dig` [cliutils.Command].
func NewCommand() cliutils.Command {
	return command{}
}

// command implements [cliutils.command].
type command struct{}

var _ cliutils.Command = command{}

//go:embed README.txt
var readme string

// Help implements [cliutils.Command].
func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

// Main implements [cliutils.Command].
func (cmd command) Main(ctx context.Context, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// 2. create an initial task to be filled according to the command line arguments
	task := &Task{
		MeasurementsWriter: os.Stdout,
		Name:               "",
		OutputWriter:       os.Stdout,
		Protocol:           "udp",
		QueryType:          "A",
		ServerAddr:         "8.8.8.8",
		ServerPort:         "53",
		URLPath:            "/dns-query",
	}

	// 3. create command line parser
	clip := pflag.NewFlagSet("rbmk dig", pflag.ContinueOnError)

	// 4. add flags to the parser
	// TODO(bassosimone): add flags to the parser

	// 5. parse command line arguments
	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
		return err
	}

	// 6. make sure we have at least one argument
	positional := clip.Args()
	if len(positional) < 1 {
		err := errors.New("missing name to resolve")
		fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
		return err
	}

	// 7. parse dig-style positional command line arguments
	var (
		countServers    int
		countQueryTypes int
	)
	for _, arg := range positional {

		// 7.1. parse the server name using the "@" syntax like in dig
		if strings.HasPrefix(arg, "@") {
			countServers++
			if countServers > 1 {
				err := fmt.Errorf("you specified more than one server to query: %s", arg)
				fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
				fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
				return err
			}
			task.ServerAddr = arg[1:]
			continue
		}

		// 7.2. parse the query options using the "+" syntax like in dig
		if strings.HasPrefix(arg, "+") {
			switch {
			case arg == "+https":
				task.Protocol = "doh"
				task.ServerPort = "443"
				continue

			case arg == "+noall":
				task.OutputWriter = io.Discard
				continue

			case arg == "+tcp":
				task.Protocol = "tcp"
				task.ServerPort = "53"
				continue

			case arg == "+tls":
				task.Protocol = "dot"
				task.ServerPort = "853"
				continue

			default:
				err := fmt.Errorf("unknown positonal argument: %s", arg)
				fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
				fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
				return err
			}
		}

		// 7.3. recognise the query type
		if _, ok := queryTypeMap[arg]; ok {
			countQueryTypes++
			if countQueryTypes > 1 {
				err := fmt.Errorf("you specified more than one query type: %s", arg)
				fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
				fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
				return err
			}
			task.QueryType = arg
			continue
		}

		// 7.4. recognise the name to resolve
		if task.Name == "" {
			task.Name = arg
			continue
		}

		// 7.5. everything else is a command line error
		err := fmt.Errorf("too many positional arguments: %s", arg)
		fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk dig --help` for usage.\n")
		return err
	}

	// 8. finally run the task
	if err := task.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk dig: %s\n", err.Error())
		return err
	}
	return nil
}
