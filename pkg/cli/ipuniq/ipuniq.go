// SPDX-License-Identifier: GPL-3.0-or-later

// Package ipuniq implements the `rbmk ipuniq` command.
package ipuniq

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/netip"
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
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk ipuniq", pflag.ContinueOnError)
	ffail := clip.BoolP("fail", "f", false, "fail on input paesing error")
	ffamily := clip.String("only", "", "only print the given address family")
	fports := clip.StringSliceP("port", "p", nil, "format output as HOST:PORT endpoints")
	frand := clip.BoolP("random", "r", false, "randomly shuffle the output")
	fromendpoints := clip.BoolP("from-endpoints", "E", false, "assume input contains endpoints")

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
			err := fmt.Errorf("invalid port number: %s", port)
			fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
			return err
		}
		ports = append(ports, uint16(value))
	}

	// 4. ensure the `--only` flag is valid
	if *ffamily != "" && *ffamily != "ipv4" && *ffamily != "ipv6" {
		err := fmt.Errorf("invalid address family: %s", *ffamily)
		fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk ipuniq --help` for usage.\n")
		return err
	}

	// 5. collect the files to read IPs from, if any. Otherwise,
	// we will read the addresses from the standard input.
	args := clip.Args()
	if len(args) <= 0 {
		args = append(args, "-")
	}

	// 6. read and parse IPs from all files
	ipAddrs := make(map[string]struct{})
	for _, fname := range args {
		if err := readIPs(env, fname, *ffail, *ffamily, *frand, *fromendpoints, ipAddrs, ports); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk ipuniq: %s\n", err.Error())
			return err
		}
	}

	// 7. if streaming, stop now
	if !*frand {
		return nil
	}

	// 8. otherwise randomly shuffle and print unique IPs w/ optional port
	var shuffled []string
	for s := range ipAddrs {
		shuffled = append(shuffled, s)
	}
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	for _, s := range shuffled {
		printAddr(env, s, ports)
	}
	return nil
}

// readIPs reads IP addresses from the given file into the given map
// and possibly streams them immediately if `frand` is false.
func readIPs(
	env cliutils.Environment,
	fname string,
	ffail bool,
	ffamily string,
	frand bool,
	fromendpoints bool,
	ipAddrs map[string]struct{},
	ports []uint16,
) error {
	var reader io.Reader
	if fname != "-" {
		filep, err := env.FS().Open(fname)
		if err != nil {
			return err
		}
		defer filep.Close()
		reader = filep
	} else {
		reader = env.Stdin()
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if fromendpoints {
			host, _, err := net.SplitHostPort(line)
			if err != nil {
				if ffail {
					return err
				}
				fmt.Fprintf(env.Stderr(), "rbmk ipuniq: warning: invalid endpoint: %s\n", line)
				continue
			}
			line = host
		}
		if ip := net.ParseIP(line); ip != nil {
			// Avoid including undesired address families in output
			parsed := netip.MustParseAddr(ip.String())
			switch {
			case ffamily == "ipv4" && !parsed.Is4():
				continue
			case ffamily == "ipv6" && !parsed.Is6():
				continue
			}

			// Implementation note: using string representation as the key to
			// handle different textual representations of same addr.
			normalized := parsed.String()
			if _, ok := ipAddrs[normalized]; ok {
				continue
			}
			ipAddrs[normalized] = struct{}{}
			if frand {
				continue
			}
			printAddr(env, normalized, ports)

		} else if ffail {
			return fmt.Errorf("invalid IP address: %s", line)
		} else {
			fmt.Fprintf(env.Stderr(), "rbmk ipuniq: warning: invalid IP address: %s\n", line)
		}
	}
	return scanner.Err()
}

// printAddr prints the address with optional port(s).
func printAddr(env cliutils.Environment, addr string, ports []uint16) {
	if len(ports) <= 0 {
		fmt.Fprintln(env.Stdout(), addr)
		return
	}
	for _, port := range ports {
		epnt := net.JoinHostPort(addr, strconv.FormatUint(uint64(port), 10))
		fmt.Fprintln(env.Stdout(), epnt)
	}
}
