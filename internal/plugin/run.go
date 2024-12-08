// SPDX-License-Identifier: GPL-3.0-or-later

package plugin

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

//go:embed run.md
var runDocs string

// newRunCommand creates the `rbmk plugin run` Command.
func newRunCommand() cliutils.Command {
	return runCommand{}
}

type runCommand struct{}

func (cmd runCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(runDocs))
	return nil
}

func (cmd runCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. if there are no arguments just print the help
	if len(argv) < 2 {
		return cmd.Help(env, argv...)
	}

	// 2. resolve the plugin binary
	binary, err := resolvePluginBinary(argv[1])
	if err != nil {
		err := fmt.Errorf("cannot resolve plugin binary: %w", err)
		fmt.Fprintf(env.Stderr(), "rbmk plugin: %s\n", err)
		return err
	}

	// 3. Ensure the RBMK_EXE environment variable is set.
	//
	// TODO(bassosimone): duplicated code
	//
	// TODO(bassosimone): unclear if we actually need this variable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("rbmk sh: cannot determine rbmk path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("rbmk sh: cannot determine absolute rbmk path: %w", err)
	}
	os.Setenv("RBMK_EXE", exePath)

	// 4. run the plugin executable
	plugin := exec.Command(binary, argv[2:]...)
	plugin.Stdin = env.Stdin()
	plugin.Stdout = env.Stdout()
	plugin.Stderr = env.Stderr()
	err = plugin.Run()

	// 5. handle the error
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk plugin: %s\n", err)
		return err
	}
	return nil
}
