// SPDX-License-Identifier: GPL-3.0-or-later

// Package ipuniq implements the `rbmk ipuniq` command.
package ipuniq

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"

	"github.com/rbmk-project/common/cliutils"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk ipuniq` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

func (cmd command) Main(ctx context.Context, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// 2. ensure we have at least one file to read
	if len(argv) < 2 {
		err := errors.New("expected one or more files containing IP addresses")
		fmt.Fprintf(os.Stderr, "rbmk ipuniq: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk ipuniq --help` for usage.\n")
		return err
	}

	// 3. read and parse IPs from all files
	ipAddrs := make(map[string]struct{})
	for _, fname := range argv[1:] {
		if err := readIPs(fname, ipAddrs); err != nil {
			fmt.Fprintf(os.Stderr, "rbmk ipuniq: %s\n", err.Error())
			return err
		}
	}

	// 4. randomly shuffle and print unique IPs
	var shuffled []string
	for s := range ipAddrs {
		shuffled = append(shuffled, s)
	}
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	for _, s := range shuffled {
		fmt.Println(s)
	}
	return nil
}

// readIPs reads IP addresses from the given file into the given map.
func readIPs(fname string, ipAddrs map[string]struct{}) error {
	filep, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer filep.Close()

	scanner := bufio.NewScanner(filep)
	for scanner.Scan() {
		line := scanner.Text()
		if ip := net.ParseIP(line); ip != nil {
			// Implementation note: using string representation as the key to
			// handle different textual representations of same addr.
			ipAddrs[ip.String()] = struct{}{}
		}
	}
	return scanner.Err()
}
