// SPDX-License-Identifier: GPL-3.0-or-later

// Package ipuniq provides the `rbmk ipuniq` Command.
package ipuniq

import (
	"bufio"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"

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
		err := errors.New("missing file operands")
		fmt.Fprintf(os.Stderr, "rbmk ipuniq: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk ipuniq --help` for usage.\n")
		return err
	}

	// 3. read and parse IPs from all files
	ips := make(map[string]struct{})
	for _, fname := range argv[1:] {
		if err := readIPs(fname, ips); err != nil {
			fmt.Fprintf(os.Stderr, "rbmk ipuniq: %s\n", err.Error())
			return err
		}
	}

	// 4. sort and print unique IPs
	var sorted []string
	for s := range ips {
		sorted = append(sorted, s)
	}
	sort.Strings(sorted)

	for _, s := range sorted {
		fmt.Println(s)
	}

	return nil
}

// readIPs reads IP addresses from a file into the provided map
func readIPs(fname string, ips map[string]struct{}) error {
	filep, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer filep.Close()

	scanner := bufio.NewScanner(filep)
	for scanner.Scan() {
		line := scanner.Text()
		if ip := net.ParseIP(line); ip != nil {
			// Note: using string representation as key to handle
			// different textual representations of same IP
			ips[ip.String()] = struct{}{}
		}
	}
	return scanner.Err()
}
