// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"testing"
)

func TestCommand(t *testing.T) {
	cmd := NewCommand()

	t.Run("help requested from the main command", func(t *testing.T) {
		cmd.Help()
	})

	t.Run("help request from the command command line", func(t *testing.T) {
		cmd.Main(context.Background(), "help")
	})

	t.Run("normal run", func(t *testing.T) {
		err := cmd.Main(context.Background(), "www.example.com")
		if err == nil || err.Error() != "not implemented" {
			t.Fatalf("expected 'not implemented', got %v", err)
		}
	})
}
