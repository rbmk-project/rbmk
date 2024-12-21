// SPDX-License-Identifier: GPL-3.0-or-later

// Package nc implements the `rbmk nc` command.
package nc

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/closepool"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

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
	clip := pflag.NewFlagSet("rbmk nc", pflag.ContinueOnError)

	// Core netcat flags (OpenBSD compatible)
	useTLS := clip.BoolP("tls", "c", false, "use TLS")
	verbose := clip.BoolP("verbose", "v", false, "verbose output")
	wait := clip.IntP("wait", "w", 0, "timeout for connect, send, and recv")
	scan := clip.BoolP("zero", "z", false, "scan for listening daemons")

	// Additional TLS features
	alpn := clip.StringSlice("alpn", nil, "TLS ALPN protocol(s)")
	options := clip.StringSliceP("option", "T", []string{}, "TLS options")
	sni := clip.String("sni", "", "TLS SNI server name")

	// RBMK specific flags
	logfile := clip.String("logs", "", "write structured logs to file")
	measure := clip.Bool("measure", false, "do not exit 1 on measurement failure")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk nc --help` for usage.\n")
		return err
	}

	// 3. validate arguments
	args := clip.Args()
	if len(args) != 2 {
		err := errors.New("expected host and port arguments")
		fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk nc --help` for usage.\n")
		return err
	}
	host, port := args[0], args[1]

	// 4. setup task with defaults
	task := &Task{
		ALPNProtocols: *alpn,
		Host:          host,
		LogsWriter:    io.Discard,
		Port:          port,
		ScanMode:      *scan,
		ServerName:    host,
		Stderr:        io.Discard,
		Stdin:         env.Stdin(),
		Stdout:        env.Stdout(),
		TLSNoVerify:   false,
		UseTLS:        *useTLS,
		WaitTimeout:   0,
	}

	// 5. finish setting up the task
	if *sni != "" {
		task.ServerName = *sni
	}
	if *wait > 0 {
		task.WaitTimeout = time.Second * time.Duration(*wait)
	}
	if *verbose {
		task.Stderr = env.Stderr()
	}
	for _, opt := range *options {
		switch opt {
		case "noverify":
			task.TLSNoVerify = true
		}
	}

	// 6. handle logs flag
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
			fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err.Error())
			return err
		}
		filepool.Add(filep)
		task.LogsWriter = io.MultiWriter(task.LogsWriter, filep)
	}

	// 7. run the task
	err := task.Run(ctx)
	if err != nil && *measure {
		fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "rbmk nc: not failing because you specified --measure\n")
		err = nil
	}

	// 8. ensure we close the opened files
	if err2 := filepool.Close(); err2 != nil {
		fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err2.Error())
		return err2
	}

	// 9. handle error when running the task
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk nc: %s\n", err.Error())
		return err
	}
	return nil
}
