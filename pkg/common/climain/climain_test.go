// SPDX-License-Identifier: GPL-3.0-or-later

package climain_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/rbmk-project/rbmk/pkg/common/climain"
	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

type fakecmd struct {
	err error
}

var _ cliutils.Command = fakecmd{}

// Help implements [cliutils.Command].
func (f fakecmd) Help(env cliutils.Environment, argv ...string) error {
	return nil
}

// Main implements [cliutils.Command].
func (f fakecmd) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	return f.err
}

func TestRun(t *testing.T) {
	t.Run("when the command does not fail", func(t *testing.T) {
		cmd := fakecmd{nil}
		climain.Run(cmd, os.Exit)
	})

	t.Run("when the command fails", func(t *testing.T) {
		var exitcode int
		cmd := fakecmd{errors.New("mocked error")}
		climain.Run(cmd, func(code int) {
			exitcode = code
		})
		if exitcode != 1 {
			t.Fatal("did not call exit")
		}
	})
}
