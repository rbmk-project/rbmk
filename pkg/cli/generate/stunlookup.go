//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// `rbmk generate stun_lookup` implementation.
//

package generate

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed stunlookup.md
var stunLookupReadme string

//go:embed stunlookup.bash
var stunLookupFunc string

// newSTUNLookupCommand creates the `rbmk generate stun_lookup` Command.
func newSTUNLookupCommand() cliutils.Command {
	return stunLookupCommand{}
}

type stunLookupCommand struct{}

// Help implements [cliutils.Command].
func (cmd stunLookupCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(stunLookupReadme))
	return nil
}

// Main implements [cliutils.Command].
func (cmd stunLookupCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk generate stun_lookup", pflag.ContinueOnError)

	// Input domains could be specified inline or read from file(s)
	inputs := clip.StringSlice("input", []string{}, "input endpoints to measure")
	inputfiles := clip.StringSlice("input-file", []string{}, "files containing input endpoints to measure")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate stun_lookup --help` for usage.\n")
		return err
	}

	// 3. ensure there are no positional arguments
	if clip.NArg() > 0 {
		err := errors.New("unexpected positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate stun_lookup --help` for usage.\n")
		return err
	}

	// 4. create the generator context
	gen := newGenerator(env.Stdout())

	// 5. write the script prefix
	gen.ScriptPrefix()
	cmd.generateLookupFunc(gen.Writer())

	// 6. compute the total number of measurements to perform
	allinputs := cmd.loadInputs(env, *inputs, *inputfiles)

	// 7. create the results directory
	gen.MakeResultsDir("stunlookup")

	// 8. generate code to measure each valid input domain
	for idx := range allinputs {
		cmd.generateMeasurementCode(gen, allinputs, idx)
	}

	// 9. compress the results directory
	gen.CompressResultsDir()

	// 10. update the progress bar
	gen.ProgressBarDone(len(allinputs))
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
func (cmd stunLookupCommand) loadInputs(
	env cliutils.Environment, inputs, inputfiles []string) (all []endpoint) {

	// 1. load the input endpoints
	for _, epnt := range loadInputs(env, "stun_lookup", inputs, inputfiles) {
		domain, port, err := net.SplitHostPort(epnt)
		if err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: warning: %s\n", err.Error())
			continue
		}
		if net.ParseIP(domain) != nil {
			fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: warning: %s is an IP address\n", domain)
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

// generateLookupFunc generates the rbmk_stun_lookup func.
func (cmd stunLookupCommand) generateLookupFunc(w io.Writer) {
	fmt.Fprintf(w, "%s\n", stunLookupFunc)
}

// generateMeasurementCode generates the shell code to measure a STUN lookup.
func (cmd stunLookupCommand) generateMeasurementCode(gen *generator, epnts []endpoint, idx int) {
	// 1. create subshell for the measurement
	epnt := epnts[idx]
	fmt.Fprintf(gen.Writer(), "# Measure %s in a subshell\n",
		net.JoinHostPort(epnt.domain, epnt.port))
	fmt.Fprintf(gen.Writer(), "(\n")

	// 2. update the progress bar
	gen.UpdateProgressBar(idx, len(epnts))

	// 3. create working directory and enter inside it
	fmt.Fprintf(gen.Writer(), "# Create working directory trying to keep the path\n")
	fmt.Fprintf(gen.Writer(), "# short to avoid Unix domain socket path issues.\n")
	fmt.Fprintf(gen.Writer(), "WORKDIR=\"$(rbmk timestamp)-$(rbmk random)\"\n")
	fmt.Fprintf(gen.Writer(), "rbmk mkdir -p \"$WORKDIR\"\n")
	fmt.Fprintf(gen.Writer(), "cd \"$WORKDIR\"\n")
	fmt.Fprintf(gen.Writer(), "\n")

	// 4. perform the measurement proper
	fmt.Fprintf(gen.Writer(), "# Note that arguments specified during\n")
	fmt.Fprintf(gen.Writer(), "# generation take precedence over the ones\n")
	fmt.Fprintf(gen.Writer(), "# passed do the script to ensure these do\n")
	fmt.Fprintf(gen.Writer(), "# not change the measured endpoints.\n")
	fmt.Fprintf(gen.Writer(), "rbmk_stun_lookup \\\n")
	fmt.Fprintf(gen.Writer(), "\t\"$@\" \\\n")
	fmt.Fprintf(gen.Writer(), "\t--stun-hostname \"%s\" \\\n", epnt.domain)
	fmt.Fprintf(gen.Writer(), "\t--stun-port \"%s\" \\\n", epnt.port)
	fmt.Fprintf(gen.Writer(), "\n")

	// 5. move the working directory once done
	fmt.Fprintf(gen.Writer(), "# Move the working directory when done\n")
	fmt.Fprintf(gen.Writer(), "cd ..\n")
	fmt.Fprintf(gen.Writer(), "rbmk mv $WORKDIR $RESULTSDIR\n")

	// 6. close the subshell
	fmt.Fprintf(gen.Writer(), ")\n")
	fmt.Fprintf(gen.Writer(), "\n")
}
