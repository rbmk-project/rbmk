// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package qa provides support for quality assurance testing of the RBMK
command line tool. The design of this package allows writing scenarios
that work both in unit/integration tests and in actual QA runs.

# Concepts

A [ScenarioEditor] is a function that modifies a [*netsim.Scenario]. We combine
editors to create complex censorship scenarios. For example, [CensorDNSLikeIran]
is an editor that implements Iran-like DNS censorship.

A [*ScenarioDescriptor] combines editors with command line arguments and
expected results to create a complete test case. The [Registry] variable
contains all the available scenarios.

The [Driver] interface abstracts running scenarios. It is compatible with
[testing.T] and testify's TestingT, allowing scenarios to be used both in
tests and standalone QA runs.

# Architecture

The package uses [github.com/rbmk-project/rbmk/pkg/x/netsim] to simulate network
conditions. Each scenario runs in an isolated network environment where
we can apply different forms of censorship and observe their effects.

Scenarios are composable: you can combine multiple editors to create
complex censorship patterns. The package provides common building blocks
(like [CensorDNSLikeIran]) that you can combine as needed.

All scenarios use a common base configuration ([MustNewCommonScenario]) that
includes frequently used hosts like dns.google and www.example.com.
*/
package qa
