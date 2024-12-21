//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Code to load inputs.
//

package generate

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/rbmk-project/common/cliutils"
)

// loadInputs loads the input domains from the command line flags.
func loadInputs(env cliutils.Environment,
	cmdname string, inputs, inputfiles []string) []string {
	// 1. prepare for input deduplication
	uniq := make(map[string]struct{})

	// 2. load the inline input domains
	for _, domain := range inputs {
		domain = strings.TrimSpace(domain) // ensure we trim spaces
		uniq[domain] = struct{}{}
	}

	// 3. load the input domains from files
	for _, filename := range inputfiles {
		for _, domain := range loadInputFile(env, cmdname, filename) {
			uniq[domain] = struct{}{}
		}
	}

	// 4. return the unique input domains
	var all []string
	for domain := range uniq {
		all = append(all, domain)
	}
	return all
}

// loadInputFile loads the input strings from the given filename.
func loadInputFile(env cliutils.Environment, cmdname, filename string) []string {
	// 1. open the file or emit warning on failure
	filep, err := env.FS().Open(filename)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate %s: warning: %s\n", cmdname, err.Error())
		return nil
	}
	defer filep.Close()

	// 2. read each input line skipping empty lines
	var inputs []string
	scanner := bufio.NewScanner(filep)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		inputs = append(inputs, line)
	}

	// 3. ensure the scanner did not encounter an error
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk generate %s: warning: %s\n", cmdname, err.Error())
		return nil
	}

	// 4. return the inputs
	return inputs
}
