//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// `rbmk generate stun_lookup` implementation.
//

package generate

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/fsx"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/rbmk-project/x/closepool"
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

	// Options controlling the output
	output := clip.String("output", "-", "output script file")
	minify := clip.Bool("minify", false, "minify the output script")

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

	// 4. collect all the input endpoints
	allinputs := cmd.loadInputs(env, *inputs, *inputfiles)

	// 5. generate the script code into a bytes buffer
	var code bytes.Buffer
	cmd.generate(allinputs, &code)

	// 6. see whether we need to open an output file
	var (
		filepool closepool.Pool
		outfp    io.Writer
	)
	switch *output {
	case "-":
		outfp = env.Stdout()
	default:
		filep, err := env.FS().OpenFile(*output, fsx.O_CREATE|fsx.O_WRONLY|fsx.O_TRUNC, 0600)
		if err != nil {
			err = fmt.Errorf("cannot open output file: %w", err)
			fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: %s\n", err.Error())
			return err
		}
		filepool.Add(filep)
		outfp = filep
	}

	// 7. format and emit the shell script
	if err := shFormatDump(code.Bytes(), *minify, outfp); err != nil {
		err = fmt.Errorf("cannot format output script: %w", err)
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: %s\n", err.Error())
		return err
	}

	// 8. ensure we close any open file successfully
	if err := filepool.Close(); err != nil {
		err = fmt.Errorf("cannot close output file: %w", err)
		fmt.Fprintf(env.Stderr(), "rbmk generate stun_lookup: %s\n", err.Error())
		return err
	}
	return nil
}

func (cmd stunLookupCommand) generate(allinputs []endpoint, w io.Writer) {
	// 1. create the generator context
	gen := newShGenerator(w)

	// 2. write the script prefix
	gen.ScriptPrefix()
	cmd.generateLookupFunc(gen.Writer())

	// 3. intercept the `-h, --help` option
	gen.WriteHelpInterceptor("rbmk_stun_lookup_help")

	// 4. create the results directory
	gen.MakeResultsDir("stunlookup")

	// 5. generate code to measure each valid input domain
	for idx := range allinputs {
		cmd.generateMeasurementCode(gen, allinputs, idx)
	}

	// 6. update the progress bar
	gen.ProgressBarDone("stunlookup", len(allinputs))

	// 7. compress the results directory
	gen.CompressResultsDir()
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
func (cmd stunLookupCommand) generateMeasurementCode(gen *shGenerator, epnts []endpoint, idx int) {
	// 1. create subshell for the measurement
	epnt := epnts[idx]
	fmt.Fprintf(gen.Writer(), "# Measure %s in a subshell\n",
		net.JoinHostPort(epnt.domain, epnt.port))
	fmt.Fprintf(gen.Writer(), "(\n")

	// 2. update the progress bar
	gen.UpdateProgressBar("stunlookup", idx, len(epnts))

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
	fmt.Fprintf(gen.Writer(), "\t--start-user-options \\\n")
	fmt.Fprintf(gen.Writer(), "\t\"$@\" \\\n")
	fmt.Fprintf(gen.Writer(), "\t--stop-user-options \\\n")
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
