// SPDX-License-Identifier: GPL-3.0-or-later

package tutorial

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

// NewCommand creates the `rbmk tutorial` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	return cmd.Main(context.Background(), env, argv...)
}

// Collect all the tutorial topics.
var (
	//go:embed basics.md
	basicsContent string

	//go:embed dns.md
	dnsContent string

	//go:embed http.md
	httpContent string

	//go:embed README.md
	readme string
)

// topicInfo contains the brief description and
// content of a given topic.
type topicInfo struct {
	brief   string
	content string
}

// topics maps tutorial topics names to their
// brief description and content.
var topics = map[string]topicInfo{
	"basics": {
		brief:   "Introduction and fundamental concepts",
		content: basicsContent,
	},

	"dns": {
		brief:   "DNS measurement patterns",
		content: dnsContent,
	},

	"http": {
		brief:   "HTTP measurement patterns",
		content: httpContent,
	},
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	switch {
	case len(argv) <= 1 || cliutils.HelpRequested(argv...):
		fmt.Fprintln(env.Stdout(), markdown.MaybeRender(readme))
		return nil

	case len(argv) > 2:
		err := fmt.Errorf("expected single tutorial topic, found: %v", argv[1:])
		fmt.Fprintf(env.Stderr(), "rbmk tutorial: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run 'rbmk tutorial' to see available topics.\n")
		return err

	default:
		topic, ok := topics[argv[1]]
		if !ok {
			err := fmt.Errorf("unknown tutorial topic: %s", argv[1])
			fmt.Fprintf(env.Stderr(), "rbmk tutorial: %s\n", err.Error())
			fmt.Fprintf(env.Stderr(), "Run 'rbmk tutorial' to see available topics.\n")
			return err
		}
		fmt.Fprintln(env.Stdout(), markdown.MaybeRender(topic.content))
		return nil
	}
}
