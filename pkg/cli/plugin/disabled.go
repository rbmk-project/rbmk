//go:build rbmk_disable_plugin

// SPDX-License-Identifier: GPL-3.0-or-later

package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
)

func newCommand() cliutils.Command {
	return disabledCommand{}
}

var _ cliutils.Command = disabledCommand{}

type disabledCommand struct{}

// Help implements [cliutils.Command].
func (d disabledCommand) Help(env cliutils.Environment, argv ...string) error {
	err := errors.New("feature disabled at compile time")
	fmt.Fprintf(env.Stderr(), "rbmk plugin: %v\n", err)
	return err
}

// Main implements [cliutils.Command].
func (d disabledCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	err := errors.New("feature disabled at compile time")
	fmt.Fprintf(env.Stderr(), "rbmk plugin: %v\n", err)
	return err
}
