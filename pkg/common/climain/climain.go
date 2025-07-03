// SPDX-License-Identifier: GPL-3.0-or-later

// Package climain implements a command's main function.
//
// You should invoke the [Run] function from the main function of your program.
package climain

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

// ExitFunc is the type of the [os.Exit] func.
type ExitFunc func(code int)

// make sure [os.Exit] implements [ExitFunc].
var _ = ExitFunc(os.Exit)

// Run runs the main function for the given command with the given [ExitFunc] and arguments.
//
// The `cmd` argument represents the command to run. We will specifically invoke the Main
// method of the [cliutils.Command] and exit (through `exitfn`) with 1 on error.
//
// The `exitfn` argument is the function to call when exiting the program, which is
// mockable so to more easily write unit tests.
//
// The `argv` arguments contain the command line arguments for the command.
//
// This function will automatically install a signal handler for [syscall.SIGINT] that
// will cancel the contect passed to [cliutils.Commnd] when receiving a signal.
func Run(cmd cliutils.Command, exitfn ExitFunc, argv ...string) {
	// 1. create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. handle signals by canceling the context
	sch := make(chan os.Signal, 1)
	signal.Notify(sch, syscall.SIGINT)
	go func() {
		defer cancel()
		<-sch
	}()

	// 3. run the selected command.
	env := cliutils.StandardEnvironment{}
	if err := cmd.Main(ctx, env, argv...); err != nil {
		exitfn(1)
	}
}
