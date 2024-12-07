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
	"strconv"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk ipuniq` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.TryRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk ipuniq", pflag.ContinueOnError)
	fports := clip.StringSliceP("port", "p", nil, "format output as HOST:PORT endpoints")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk ipuniq --help` for usage.\n")
		return err
	}

	// 3. ensures all the ports are valid
	var ports []uint16
	for _, port := range *fports {
		value, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
			return err
		}
		ports = append(ports, uint16(value))
	}

	// 4. ensure we have at least one file to read
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("expected one or more files containing IP addresses")
		fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk ipuniq --help` for usage.\n")
		return err
	}

	// 5. read and parse IPs from all files
	ipAddrs := make(map[string]struct{})
	for _, fname := range args {
		if err := readIPs(fname, ipAddrs); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
			return err
		}
	}

	// 6. randomly shuffle and print unique IPs w/ optional port
	var shuffled []string
	for s := range ipAddrs {
		shuffled = append(shuffled, s)
	}
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	for _, s := range shuffled {
		if len(ports) <= 0 {
			fmt.Fprintln(env.Stdout(), s)
			continue
		}
		for _, port := range ports {
			epnt := net.JoinHostPort(s, strconv.FormatUint(uint64(port), 10))
			fmt.Fprintln(env.Stdout(), epnt)
		}
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
