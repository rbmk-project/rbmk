// SPDX-License-Identifier: GPL-3.0-or-later

package cliutils_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rbmk-project/rbmk/cmd/internal/cliutils"
)

type fakecmd struct {
	err error
}

var _ cliutils.Command = fakecmd{}

// Help implements [cliutils.Command].
func (f fakecmd) Help(argv ...string) error {
	return nil
}

// Main implements [cliutils.Command].
func (f fakecmd) Main(ctx context.Context, argv ...string) error {
	return f.err
}

func TestCommandWithSubCommands(t *testing.T) {
	type testcase struct {
		argv    []string
		failure string
	}

	cases := []testcase{{
		argv:    []string{"buresu"},
		failure: "",
	}, {
		argv:    []string{"buresu", "-h"},
		failure: "",
	}, {
		argv:    []string{"buresu", "--help"},
		failure: "",
	}, {
		argv:    []string{"buresu", "help"},
		failure: "",
	}, {
		argv:    []string{"buresu", "help", "env"},
		failure: "",
	}, {
		argv:    []string{"buresu", "help", "__nonexistent__"},
		failure: "no such help topic",
	}, {
		argv:    []string{"buresu", "env"},
		failure: "",
	}, {
		argv:    []string{"buresu", "__nonexistent__"},
		failure: "no such command",
	}}

	for _, tc := range cases {
		t.Run(strings.Join(tc.argv, " "), func(t *testing.T) {
			cmd := cliutils.NewCommandWithSubCommands(
				"rbmk", "", map[string]cliutils.Command{
					"env": fakecmd{},
				},
			)
			err := cmd.Main(context.Background(), tc.argv...)
			switch {
			case tc.failure == "" && err == nil:
				// all good
			case tc.failure != "" && err == nil:
				t.Fatal("expected", tc.failure, "got", err)
			case tc.failure == "" && err != nil:
				t.Fatal("expected", tc.failure, "got", err.Error())
			case tc.failure != "" && err != nil:
				if err.Error() != tc.failure {
					t.Fatal("expected", tc.failure, "got", err.Error())
				}
			}
		})
	}
}

func TestHelpRequested(t *testing.T) {
	type testcase struct {
		argv   []string
		expect bool
	}

	cases := []testcase{
		{
			argv:   []string{"--log-file", "log.jsonl", "-h"},
			expect: true,
		},

		{
			argv:   []string{"--log-file", "log.jsonl", "help"},
			expect: true,
		},

		{
			argv:   []string{"--log-file", "log.jsonl", "--help"},
			expect: true,
		},

		{
			argv:   []string{"--log-file", "log.jsonl"},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(strings.Join(tc.argv, " "), func(t *testing.T) {
			got := cliutils.HelpRequested(tc.argv...)
			if got != tc.expect {
				t.Fatal("expected", tc.expect, "got", got)
			}
		})
	}
}
