// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"testing"

	"github.com/rbmk-project/rbmk/pkg/common/cliutils"
)

func TestCommand(t *testing.T) {
	stdenv := cliutils.StandardEnvironment{}
	cmd := NewCommand()

	t.Run("help requested from the main command", func(t *testing.T) {
		cmd.Help(stdenv)
	})

	t.Run("help request from the command command line", func(t *testing.T) {
		cmd.Main(context.Background(), stdenv, "help")
	})

	t.Run("normal run", func(t *testing.T) {
		err := cmd.Main(context.Background(), stdenv, "dig")
		if err == nil || err.Error() != "missing name to resolve" {
			t.Fatalf("expected 'not implemented', got %v", err)
		}
	})
}
