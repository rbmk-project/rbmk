// SPDX-License-Identifier: GPL-3.0-or-later

package intro

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/rbmk-project/common/cliutils"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk intro` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(argv ...string) error {
	return cmd.Main(context.Background(), argv...)
}

func (cmd command) Main(ctx context.Context, argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}
