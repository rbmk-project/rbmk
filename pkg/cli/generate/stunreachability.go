//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// `rbmk generate stun_reachability` implementation.
//

package generate

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var stunReachabilityReadme string

//go:embed stunreachability.bash
var stunReachabilityFunc string

// newSTUNReachabilityCommand creates the `rbmk generate stun_reachability` Command.
func newSTUNReachabilityCommand() cliutils.Command {
	return stunReachabilityCommand{}
}

type stunReachabilityCommand struct{}

// Help implements [cliutils.Command].
func (cmd stunReachabilityCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(stunReachabilityReadme))
	return nil
}

// Main implements [cliutils.Command].
func (cmd stunReachabilityCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk generate stun_reachability", pflag.ContinueOnError)

	// Input domains could be specified inline or read from file(s)
	inputs := clip.StringSlice("input", []string{}, "input endpoints to measure")
	inputfiles := clip.StringSlice("input-file", []string{}, "files containing input endpoints to measure")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_reachability: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate stun_reachability --help` for usage.\n")
		return err
	}

	// 3. ensure there are no positional arguments
	if clip.NArg() > 0 {
		err := errors.New("unexpected positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_reachability: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate stun_reachability --help` for usage.\n")
		return err
	}

	// 4. create the generator context
	generate := newGenerator(env.Stdout())

	// 5. write the script prefix
	generate.ScriptPrefix()
	generate.StunReachabilityFunc()

	// 6. create the results directory
	generate.MakeResultsDir("stunreachability")

	// 7. compute the total number of measurements to perform
	allinputs := cmd.loadInputs(env, *inputs, *inputfiles)
	generate.InitializeProgressBar(len(allinputs))

	// 8. generate code to measure each valid input domain
	for _, epnt := range allinputs {
		// 8.1. start the measurement
		generate.StartSTUNReachability(epnt)

		// 8.2. update the progress bar
		generate.UpdateProgressBar()
	}

	// 9. compress the measurement results
	generate.CompressWorkDir()
	return nil
}

// endpoint is a network endpoint.
type endpoint struct {
	// domain is the domain name.
	domain string

	// port is the port number.
	port string
}

// loadInputs loads the input domains from the command line flags.
func (cmd stunReachabilityCommand) loadInputs(
	env cliutils.Environment, inputs, inputfiles []string) (all []endpoint) {

	// 1. load the input endpoints
	for _, epnt := range loadInputs(env, "stun_reachability", inputs, inputfiles) {
		domain, port, err := net.SplitHostPort(epnt)
		if err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk generate stun_reachability: warning: %s\n", err.Error())
			continue
		}
		if net.ParseIP(domain) != nil {
			fmt.Fprintf(env.Stderr(), "rbmk generate stun_reachability: warning: %s is an IP address\n", domain)
			continue
		}
		all = append(all, endpoint{domain, port})
	}

	// 2. if empty, generate default endpoints
	//
	// We should sync with https://gitlab.torproject.org/tpo/applications/tor-browser-build/-/blob/main/projects/tor-expert-bundle/pt_config.json
	if len(all) <= 0 {
		all = append(all, endpoint{"stun.l.google.com", "19302"})
		all = append(all, endpoint{"stun.antisip.com", "3478"})
		all = append(all, endpoint{"stun.bluesip.net", "3478"})
		all = append(all, endpoint{"stun.dus.net", "3478"})
		all = append(all, endpoint{"stun.epygi.com", "3478"})
		all = append(all, endpoint{"stun.sonetel.com", "3478"})
		all = append(all, endpoint{"stun.uls.co.za", "3478"})
		all = append(all, endpoint{"stun.voipgate.com", "3478"})
		all = append(all, endpoint{"stun.voys.nl", "3478"})
	}
	return
}

/*
StartSTUNReachability generates the shell code to measure STUN reachability.

Arguments:

1. endpoint is the STUN endpoint to measure.

Notes:

1. Assumes that the WORKDIR variable is already defined.

2. Assumes that the `rbmk_measure_stun_reachability` shell function is available.

3. Assumes that the `rbmk_output_format_dir_prefix` shell function is available.

4. The measurement is performed in the foreground.
*/
func (g *generator) StartSTUNReachability(epnt endpoint) {
	fmt.Fprintf(g.writer, "# Measure %s\n", epnt)
	fmt.Fprintf(g.writer, "rbmk_measure_stun_reachability \\\n")
	fmt.Fprintf(
		g.writer,
		"\t\"$WORKDIR/$(rbmk_output_format_dir_prefix $COUNT)-stunreachability\" \\\n",
	)
	fmt.Fprintf(g.writer, "\t\"%s\" \"%s\"\n", epnt.domain, epnt.port)
	fmt.Fprintf(g.writer, "\n")
}

// StunReachabilityFunc generates the `rbmk_measure_stun_reachability` shell function.
func (g *generator) StunReachabilityFunc() {
	fmt.Fprintf(g.writer, "%s\n", stunReachabilityFunc)
}
