// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package censor implements common internet censorship techniques for testing.

This package provides filters that model different types of censorship:

All filters implement the [packet.Filter] interface and can be composed to model
complex censorship scenarios. The implementations focus on common real-world
censorship techniques while remaining simple enough for testing purposes.

# DNS Response Injection

The [*DNSPoisoner] type implements GFW-style DNS poisoning by injecting spoofed
responses. It can target specific resolvers and is based on a database of poisoned
responses to inject. Legitimate responses are allowed to pass through, thus the
client is expected to receive multiple responses for each censored query.

# TCP Reset Injection

The [*TCPResetter] type implements RST-based connection disruption. It can match
on specific patterns (e.g., TLS SNI) while allowing TCP handshakes to complete,
modeling how real censors selectively terminate connections based on application
layer content. Combining pattern matching and endpoint matching allows for modeling
SNI+endpoint based blocking, which is another common censorship case.

# Connection Blackholing

The [*Blackholer] type implements connection blackholing with optional pattern
matching. Once triggered, it blocks all packets for the matching connection
for a configurable duration, causing residual five-tuple censorship. This models
censors that completely block specific traffic patterns or endpoints, causing
I/O timeouts.

# Destination NAT

The [*DNatter] type implements transparent proxying through destination NAT
(DNAT): it allows redirecting traffic from specific sources to alternative destinations
while maintaining proper connection tracking. This models censors that redirect
traffic to warning pages or surveillance systems.
*/
package censor
