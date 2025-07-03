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

Go 1.23.

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

Core Measurement Commands:
- `curl`: Measures HTTP/HTTPS endpoints with `curl(1)`-like syntax.
- `dig`: Performs DNS measurements with `dig(1)`-like syntax.
- `nc`: Measures TCP and TLS endpoints with an OpenBSD `nc(1)`-like syntax.
- `stun`: Resolves the public IP addresses using STUN.

Unix-like Commands for Scripting:
- `cat`: Concatenates files.
- `head`: Print first lines of files.
- `ipuniq`: Shuffle, deduplicate, and format IP addresses.
- `markdown`: Renders Markdown to console.
- `mkdir`: Creates directories.
- `mv`: Moves (renames) files and directories.
- `pipe`: Creates named pipes for inter-process communication.
- `random`: Generates random bytes.
- `rm`: Removes files and directories.
- `sh`: Runs POSIX shell scripts.
- `tar`: Creates tar archives.
- `timestamp`: Prints filesystem-friendly timestamps.
- `version`: Prints the `rbmk` version.

Helper Commands:
- `intro`: Shows a brief introduction with usage examples.
- `tutorial`: Provides comprehensive usage documentation.

Each command supports the `--help` flag for detailed usage information.

## Release Builds

You need GNU make installed. Run:

```bash
make release
```

Run `make` without arguments to see all available targets.

## Documentation

Read the packages documentation at [pkg.go.dev/rbmk-project/rbmk](
https://pkg.go.dev/github.com/rbmk-project/rbmk).

## Design

The [docs/design](./docs/design) directory contains all the design documents.

## Contributing

Contributions are welcome! Please submit pull requests using GitHub.

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```
