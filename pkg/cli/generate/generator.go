//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Code to generate shell scripts.
//

package generate

import (
	"fmt"
	"io"
)

// generator generates a shell script.
//
// The zero value IS NOT ready to use and you must explicitly
// call [newGenerator] to obtain a ready-to-use generator.
type generator struct {
	// stunIPv4 is the IPv4 address of the STUN server or empty
	// if the STUN functionality is disabled for IPv4.
	stunIPv4 string

	// stunIPv6 is the IPv6 address of the STUN server or empty
	// if the STUN functionality is disabled for IPv6.
	stunIPv6 string

	// stunPort is the port of the STUN server.
	stunPort string

	// writer is the writer to which we write the shell script.
	writer io.Writer
}

// newGenerator creates a new [*generator] using the given [io.Writer]
// as well as the given configuration. Note that stunIPv4 and/or stunIPv6
// may be empty if the STUN functionality is disabled.
func newGenerator(
	w io.Writer, stunIPv4, stunIPv6, stunPort string) *generator {
	return &generator{
		stunIPv4: stunIPv4,
		stunIPv6: stunIPv6,
		stunPort: stunPort,
		writer:   w,
	}
}

/*
CompressWorkDir generates the code to compress the working
directory into a tarball and remove the original directory.

Notes:

1. Assumes that the WORKDIR variable is already defined.
*/
func (g *generator) CompressWorkDir() {
	fmt.Fprintf(g.writer, "# Compress the working directory\n")
	fmt.Fprintf(g.writer, "rbmk tar -czf \"$WORKDIR.tar.gz\" \"$WORKDIR\"\n")
	fmt.Fprintf(g.writer, "rbmk rm -rf \"$WORKDIR\"\n")
	fmt.Fprintf(g.writer, "\n")
}

// InitializeProgressBar generates the code
// to initialize the progress bar.
func (g *generator) InitializeProgressBar(total int) {
	fmt.Fprintf(g.writer, "# Initialize the progress bar\n")
	fmt.Fprintf(g.writer, "TOTAL=%d\n", total)
	fmt.Fprintf(g.writer, "COUNT=0\n")
	fmt.Fprintf(g.writer, "\n")
}

// MakeResultsDir generates the code to
// make the results directory.
func (g *generator) MakeResultsDir(name string) {
	fmt.Fprintf(g.writer, "# Make the results directory\n")
	fmt.Fprintf(g.writer, "WORKDIR=\"$(rbmk timestamp)-%s\"\n", name)
	fmt.Fprintf(g.writer, "rbmk mkdir -p \"$WORKDIR\"\n")
	fmt.Fprintf(g.writer, "\n")
}

/*
MaybeStartSTUN generates code to perform STUN lookups if the
corresponding IP addresses and port have been set, and the
measurement index is zero or a multiple of 8.

Assumptions:

1. `$WORKDIR` exists and contains the working directory.

2. `rbmk_stun_lookup` is an existing shell function
that performs a STUN lookup.

3. `rbmk_output_format_dir_prefix` is an existing shell
function that generates a directory prefix.

Notes:

1. STUN lookups do not increment the progress bar and may
occur multiple times during a given measurement.

2. STUN lookups run in the backgroun and the caller must
eventually invoke `wait` to synchronise on their completion.
*/
func (g *generator) MaybeStartSTUN(index int, stunDomain string) {
	// 1. Do nothing if the index is not zero or a multiple of 8
	if index%8 != 0 {
		return
	}

	// 2. Generate code for STUN IPv4 lookup
	if g.stunIPv4 != "" && g.stunPort != "" {
		fmt.Fprintf(g.writer, "# Resolve the reflexive IPv4 address using STUN.\n")
		fmt.Fprintf(g.writer, "rbmk_stun_lookup \\\n")
		fmt.Fprintf(
			g.writer,
			"\t\"$WORKDIR/$(rbmk_output_format_dir_prefix $COUNT)-stun-v4-%s\" \\\n",
			stunDomain,
		)
		fmt.Fprintf(g.writer, "\t\"%s\" \"%s\" &\n", g.stunIPv4, g.stunPort)
		fmt.Fprintf(g.writer, "\n")
	}

	// 3. Generate code for STUN IPv6 lookup
	if g.stunIPv4 != "" && g.stunPort != "" {
		fmt.Fprintf(g.writer, "# Resolve the reflexive IPv6 address using STUN.\n")
		fmt.Fprintf(g.writer, "rbmk_stun_lookup \\\n")
		fmt.Fprintf(
			g.writer,
			"\t\"$WORKDIR/$(rbmk_output_format_dir_prefix $COUNT)-stun-v6-%s\" \\\n",
			stunDomain,
		)
		fmt.Fprintf(g.writer, "\t\"%s\" \"%s\" &\n", g.stunIPv6, g.stunPort)
		fmt.Fprintf(g.writer, "\n")
	}
}

// ScriptPrefix generates the shell script prefix including
// all the necessary built-in shell functions and variables.
func (g *generator) ScriptPrefix() {
	fmt.Fprintf(g.writer, "%s\n", prefixBash)
}

// WaitForMeasurements generates the code to wait
// for background measurements to finish.
func (g *generator) WaitForMeasurements() {
	fmt.Fprintf(g.writer, "# Wait for measurements to finish\n")
	fmt.Fprintf(g.writer, "wait\n")
	fmt.Fprintf(g.writer, "\n")
}

/*
UpdateProgressBar generates the code to update the progress bar.

Notes:

1. Assumes that the COUNT and TOTAL variables are already defined.

2. Assumes that the `rbmk_ui_print_progress` shell function is available.
*/
func (g *generator) UpdateProgressBar() {
	fmt.Fprintf(g.writer, "# Update the progress bar\n")
	fmt.Fprintf(g.writer, "COUNT=$((COUNT+1))\n")
	fmt.Fprintf(g.writer, "rbmk_ui_print_progress $COUNT $TOTAL\n")
	fmt.Fprintf(g.writer, "\n")
}
