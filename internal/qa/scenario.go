// SPDX-License-Identifier: GPL-3.0-or-later

package qa

import (
	"context"

	"github.com/rbmk-project/rbmk/internal/cli"
	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/rbmk-project/x/netsim"
	"github.com/rbmk-project/x/netsim/geolink"
	"github.com/stretchr/testify/require"
)

// ScenarioEditor modifies a [*netsim.Scenario]. Editors are the building
// blocks used to create complex censorship scenarios. They are composable:
// you can combine multiple editors to create sophisticated test cases.
type ScenarioEditor func(scenario *netsim.Scenario) *netsim.Scenario

// MustNewCommonScenario creates a new netsim.Scenario with frequently
// used hosts (dns.google, www.example.com) already configured. It
// panics on any initialization error.
//
// The cacheDir parameter specifies where to cache TLS certificates.
func MustNewCommonScenario(cacheDir string) *netsim.Scenario {
	scenario := netsim.NewScenario(cacheDir)
	scenario.Attach(scenario.MustNewGoogleDNSStack())
	scenario.Attach(scenario.MustNewExampleComStack())
	return scenario
}

// ScenarioDescriptor describes a complete test scenario, including the
// network conditions to simulate (via Editors), the command to run (via
// Argv), and the expected outcome (via ExpectedErr).
type ScenarioDescriptor struct {
	// Name identifies this scenario for logging and debugging.
	Name string

	// Editors modify the network environment for this scenario.
	Editors []ScenarioEditor

	// Argv contains the command line arguments to execute.
	Argv []string

	// ExpectedErr is the error we expect from running
	// the command. If nil, we expect the command to succeed.
	ExpectedErr error
}

// Run runs the given [*ScenarioDescriptor] using the given [*Driver].
func (desc *ScenarioDescriptor) Run(t Driver) {
	// Initialize the scenario and apply all the editors.
	scenario := MustNewCommonScenario("testdata")
	defer scenario.Close()
	for _, modifier := range desc.Editors {
		scenario = modifier(scenario)
	}

	// Obtain the client stack and override the function used
	// to dial new network connections, to use the simulated stack
	// rather than using the host's network stack.
	stack := scenario.MustNewClientStack()
	linkConfig := &geolink.Config{
		Delay: 0, // TODO(bassosimone): set delay? make configurable?
		Log:   true,
	}
	scenario.Attach(geolink.Extend(stack, linkConfig))
	testable.DialContext.Set(stack.DialContext)

	// Create the main RBMK command.
	cmd := cli.NewCommand()

	// Execute the given argv.
	err := cmd.Main(context.Background(), desc.Argv...)

	// Check whether the return value is OK.
	if desc.ExpectedErr != nil {
		require.EqualError(t, err, desc.ExpectedErr.Error(),
			"scenario %s should return expected error", desc.Name)
	} else {
		require.NoError(t, err, "scenario %s should not return error", desc.Name)
	}
}
