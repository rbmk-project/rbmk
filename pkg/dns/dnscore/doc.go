// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package dnscore provides a DNS resolver, a DNS transport, a query builder,
and a DNS response parser.

This package is designed to facilitate DNS measurements and queries
by providing both high-level and low-level APIs. It aims to be flexible,
extensible, and easy to integrate with existing Go code.

The high-level [*Resolver] API provides a DNS resolver that is compatible with the
[*net.Resolver] struct from the [net] package. The low-level [*Transport] API
allows users to send and receive DNS messages using different protocols
and dialers. The package also includes utilities for creating and validating
DNS messages.

# Features

- High-level [*Resolver] API compatible with [*net.Resolver] for easy integration.

- Low-level [*Transport] API allowing granular control over DNS requests and responses.

- Support for multiple DNS protocols, including UDP, TCP, DoT, DoH, and DoQ.

- Utilities for creating and validating DNS messages.

- Optional logging for structured diagnostic events through [log/slog].

- Handling of duplicate responses for DNS over UDP to measure censorship.

The package is structured to allow users to compose their own workflows
by providing building blocks for DNS queries and responses. It uses
the widely-used [github.com/miekg/dns] library for DNS message parsing
and serialization.

# Design Documents

The [dd-000-dnscore] document describes the design of this package.

The [df-000-dns] document describes the data format generated by this
package when using [log/slog] to emit structured diagnostic events.

[dd-000-dnscore]: https://rbmk-project.github.io/rbmk/design/dd-000-dnscore
[df-000-dns]: https://rbmk-project.github.io/rbmk/spec/data-format/df-000-dns
*/
package dnscore
