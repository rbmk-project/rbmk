//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// `rbmk generate sni_blocking` implementation.
//

package generate

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var sniBlockingReadme string

// newSNIBlockingCommand creates the `rbmk generate sni_blocking` Command.
func newSNIBlockingCommand() cliutils.Command {
	return sniBlockingCommand{}
}

type sniBlockingCommand struct{}

// Help implements [cliutils.Command].
func (cmd sniBlockingCommand) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(sniBlockingReadme))
	return nil
}

// Main implements [cliutils.Command].
func (cmd sniBlockingCommand) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk generate sni_blocking", pflag.ContinueOnError)

	// IPv4 and IPv6 are enabled by default
	disableIPv4 := clip.Bool("disable-ipv4", false, "disable measuring IPv4 endpoints")
	disableIPv6 := clip.Bool("disable-ipv6", false, "disable measuring IPv6 endpoints")

	// STUN lookups, instead, are disabled by default
	enableSTUN := clip.Bool("enable-stun", false, "enable STUN lookups")
	stunDomain := clip.String("stun-domain", "stun.l.google.com", "STUN server domain")
	stunPort := clip.String("stun-port", "19302", "STUN server port")

	// Input domains could be specified inline or read from file(s)
	inputs := clip.StringSlice("input", []string{}, "input domains to measure")
	inputfiles := clip.StringSlice("input-file", []string{}, "files containing input domains to measure")

	// By default we use `www.example.com` as the test helper domain since it's
	// likely to be working and not censored in most places
	testHelperDomain := clip.String("test-helper-domain", "www.example.com", "test helper domain")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate sni_blocking: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate sni_blocking --help` for usage.\n")
		return err
	}

	// 3. ensure there are no positional arguments
	if clip.NArg() > 0 {
		err := errors.New("unexpected positional arguments")
		fmt.Fprintf(env.Stderr(), "rbmk generate sni_blocking: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk generate sni_blocking --help` for usage.\n")
		return err
	}

	// 4. create the discoverer context and discover the test helper
	// addresses as well as the STUN server addresses, if needed.
	discover := newDiscoverer(*disableIPv4, *disableIPv6, *enableSTUN)
	stunAddrV4, stunAddrV6, err := discover.STUNAddrs(ctx, *stunDomain)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate sni_blocking: %s\n", err.Error())
		return err
	}
	thAddrV4, thAddrV6, err := discover.TestHelperAddrs(ctx, *testHelperDomain)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate sni_blocking: %s\n", err.Error())
		return err
	}

	// 5. create the generator context
	generate := newGenerator(
		env.Stdout(),
		stunAddrV4,
		stunAddrV6,
		*stunPort,
	)

	// 6. write the script prefix
	generate.ScriptPrefix()

	// 7. create the results directory
	generate.MakeResultsDir("sniblocking")

	// 8. compute the total number of measurements to perform
	allinputs := cmd.loadInputs(env, *inputs, *inputfiles)
	generate.InitializeProgressBar(len(allinputs))

	// 9. generate code to measure each valid input domain
	for index, domain := range allinputs {
		// 9.1. periodically refresh STUN information
		generate.MaybeStartSTUN(index, *stunDomain)

		// 9.2. possibly measure the domain using IPv4
		generate.MaybeStartSNIBlocking(domain, "v4", thAddrV4)

		// 9.3. possibly measure the domain using IPv6
		generate.MaybeStartSNIBlocking(domain, "v6", thAddrV6)

		// 9.4. wait for measurements to finish
		generate.WaitForMeasurements()

		// 9.5. update the progress bar
		generate.UpdateProgressBar()
	}

	// 10. compress the measurement results
	generate.CompressWorkDir()
	return nil
}

// loadInputs loads the input domains from the command line flags.
func (cmd sniBlockingCommand) loadInputs(
	env cliutils.Environment, inputs, inputfiles []string) (all []string) {
	for _, domain := range loadInputs(env, "sni_blocking", inputs, inputfiles) {
		if !domainRegexp.MatchString(domain) {
			fmt.Fprintf(env.Stderr(), "rbmk generate sni_blocking: warning: invalid domain: %s\n", domain)
			continue
		}
		all = append(all, domain)
	}
	return
}

/*
MaybeStartSNIBlocking generates the shell code to measure SNI blocking.

Arguments:

1. domain is the domain to measure.

2. family is the IP family to use (either "v4" or "v6").

3. addr is the IP address to use or the empty string if the
specific address family is disabled.

Notes:

1. Assumes that the WORKDIR variable is already defined.

2. Assumes that the `rbmk_measure_sni_blocking` shell function is available.

3. Assumes that the `rbmk_output_format_dir_prefix` shell function is available.

4. The measurement is performed in the background.
*/
func (g *generator) MaybeStartSNIBlocking(domain string, family string, addr string) {
	if addr == "" {
		return
	}
	fmt.Fprintf(g.writer, "# Measure %s using %s\n", domain, addr)
	fmt.Fprintf(g.writer, "rbmk_measure_sni_blocking \\\n")
	fmt.Fprintf(
		g.writer,
		"\t\"$WORKDIR/$(rbmk_output_format_dir_prefix $COUNT)-sniblocking-%s-%s\" \\\n",
		family,
		domain,
	)
	fmt.Fprintf(g.writer, "\t\"%s\" 443 \"%s\" &\n", addr, domain)
	fmt.Fprintf(g.writer, "\n")
}
