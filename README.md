# Really Basic Measurement Kit (RBMK)

[![GoDoc](https://pkg.go.dev/badge/github.com/rbmk-project/rbmk)](https://pkg.go.dev/github.com/rbmk-project/rbmk) [![Build Status](https://github.com/rbmk-project/rbmk/actions/workflows/go.yml/badge.svg)](https://github.com/rbmk-project/rbmk/actions) [![codecov](https://codecov.io/gh/rbmk-project/rbmk/branch/main/graph/badge.svg)](https://codecov.io/gh/rbmk-project/rbmk)

RBMK is a CLI tool for performing low-level, scriptable network measurements.

## Features

1. Run commands you already know (e.g., `rbmk dig`, `rbmk curl`).

2. Measure fundamental network operations: DNS, HTTP(S), TCP, TLS, and STUN.

3. Obtain structured logs in JSONL format for easy analysis using `--logs FILE`.

4. Organize measurements as portable shell scripts run using `rbmk sh`.

5. Support shell scripting with built-in commands like `rbmk tar`, and `rbmk mv`.

6. Get extensive help using `rbmk help` and `rbmk tutorial`.

## Use Cases

RBMK is mostly useful when investigating network anomalies, including
outages, misconfigurations, censorship, and performance issues.

## Minimum Required Go Version

Go 1.25

## Installation

```bash
go install -v -tags netgo github.com/rbmk-project/rbmk/cmd/rbmk@latest
```

## Quick Start

These examples demonstrate how to use RBMK for common network measurements:

```bash
# Resolve a domain name
rbmk dig +short=ip example.com
93.184.215.14

# Make an HTTP request
rbmk curl -vo index.html https://example.com/

# Combine dig and curl for step-by-step measurement
addr=$(rbmk dig +short=ip example.com | rbmk head -n 1)
rbmk curl --resolve example.com:443:$addr https://example.com/

# Use --logs to get structured logs in JSONL format
rbmk dig --logs dns.jsonl +short=ip example.com
rbmk curl --logs http.jsonl -vo index.html https://example.com/
```

For a quick introduction with more examples, run:

```sh
rbmk intro
```

For comprehensive usage documentation, run:

```sh
rbmk tutorial
```

## Build Tags

RBMK supports the following build tags to customize the build:

| Feature Flag            | Description                                                     |
| ----------------------- | --------------------------------------------------------------- |
| `netgo`                 | Use pure-Go functions instead of linking the C stdlib.          |
| `rbmk_disable_markdown` | Disables Markdown rendering in help text, reducing binary size. |

You can pass those flags to `go install` or `go build`. For example:

```bash
go install -v -tags netgo,rbmk_disable_markdown github.com/rbmk-project/rbmk/cmd/rbmk@latest
```

## Commands

1. Core Measurement Commands: `curl`, `dig`, `nc`, `stun`.

2. Unix-like Commands for Scripting: `cat`, `head`, `mkdir`, `mv`, `rm`, `sh`, `tar`.

3. RBMK-Specific Commands for Scripting: `ipuniq`, `markdown`, `pipe`, `random`, `timestamp`.

4. Helper Commands: `intro`, `tutorial`, `version`.

Each command supports the `--help` flag for detailed usage information.

For example:

```bash
rbmk curl --help
```

## Release Builds

You need GNU make installed. Run:

```bash
make release
```

Run `make` without arguments to see all available targets.

## Documentation

Read the packages documentation at [pkg.go.dev/rbmk-project/rbmk](https://pkg.go.dev/github.com/rbmk-project/rbmk).

## Architecture

**Documentation:**
- [docs/design](./docs/design): Design documents.
- [docs/man](./docs/man): Manual pages for RBMK commands.
- [docs/spec](./docs/spec): Specification documents.
- [docs/tutorial](./docs/tutorial): Tutorials for using RBMK.

**Main Entry Point:**
- [cmd/rbmk](./cmd/rbmk): The main RBMK command-line tool.

**Go Packages:**
- [pkg/cli](./pkg/cli): CLI implementation.
- [pkg/common](./pkg/common): Common utilities and helpers.
- [pkg/dns](./pkg/dns): DNS measurement implementation.
- [pkg/x](./pkg/x): Experimental Go packages.

**Build System:**
- [GNUmakefile](./GNUmakefile): Makefile for RBMK.

## Contributing

Contributions are welcome! Please submit pull requests using GitHub.

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```
