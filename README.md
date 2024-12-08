# Really Basic Measurement Kit

[![GoDoc](https://pkg.go.dev/badge/github.com/rbmk-project/rbmk)](https://pkg.go.dev/github.com/rbmk-project/rbmk)

RBMK (Really Basic Measurement Kit) is a command-line utility
to facilitate network exploration and measurements. It provides
fundamental network operations (`dig`, `curl`, and `stun`) that you can
compose together to perform modular network measurements where
you can observe each operation in isolation.

## Features

- Modular design with DNS, HTTP(S), and STUN measurement operations
- CLI-first approach with composable subcommands
- Extensive structured logging for detailed analysis
- Support for multiple DNS protocols (UDP, TCP, DoT, and DoH)
- Integrated online help with optional markdown rendering

- Core Measurement Commands:
  - `dig`: DNS measurements with multiple protocols
  - `curl`: HTTP(S) endpoint measurements
  - `stun`: Resolve the public IP addresses

- Scripting Support:
  - Built-in POSIX shell interpreter
  - Cross-platform Unix-like subcommands (e.g., `cat`, `tar`, `mv`)
  - Script generation tools

The tool is designed to support both general use and measurement-specific
features, with support for scripting concurrent operations and
extensive integration testing capabilities.

## Minimum Go version

Go 1.23

## Installation

```sh
go install github.com/rbmk-project/rbmk/cmd/rbmk@latest
```

## Building

```sh
go build -v ./cmd/rbmk
```

If you have GNU make installed, you can also run:

```sh
make
```

to see all the available build/install options.

## Feature Flags

We support the following build-time feature flags:

* `rbmk_disable_markdown` disables markdown rendering when
producing help text thus making the binary much smaller.

You need to pass these feature flags to the `go build` command
or the `go install` command using the `-tags` flag.

For example, this command:

```sh
go build -v -tags rbmk_disable_markdown,netgo ./cmd/rbmk
```

builds with disabled markdown rendering (`rbmk_disable_markdown`) and
using the pure-Go DNS lookup engine (`netgo`).

## Quick Start

```sh
# Resolve a domain name
$ rbmk dig +short=ip example.com
93.184.215.14

# Make an HTTP request
$ rbmk curl https://example.com/
...

# Combine dig and curl for step-by-step measurement
$ IP=$(rbmk dig +short=ip example.com|head -n1)
$ rbmk curl --resolve example.com:443:$IP https://example.com/

# Collect measurement data in flat JSONL format
$ rbmk dig --logs dns.jsonl example.com
$ rbmk curl --logs http.jsonl https://example.com/
```

For a quick introduction with more examples, run:

```sh
$ rbmk intro
```

For comprehensive usage documentation, run:

```sh
$ rbmk tutorial
```

## Commands

Core Measurement Commands:
- `curl`: Measures HTTP/HTTPS endpoints with `curl(1)`-like syntax.
- `dig`: Performs DNS measurements with `dig(1)`-like syntax.
- `stun`: Resolves the public IP addresses using STUN.

Unix-like Commands for Scripting:
- `cat`: Concatenates files.
- `ipuniq`: Shuffle, deduplicate, and format IP addresses.
- `mkdir`: Creates directories.
- `mv`: Moves (renames) files and directories.
- `rm`: Removes files and directories.
- `sh`: Runs POSIX shell scripts.
- `tar`: Creates tar archives.
- `timestamp`: Prints filesystem-friendly timestamps.

Helper Commands:
- `intro`: Shows a brief introduction with usage examples.
- `tutorial`: Provides comprehensive usage documentation.

Each command supports the `--help` flag for detailed usage information.

## Design

The project focuses on modular, composable measurements where each
operation that may fail is executed independently. This allows for precise
analysis of network behavior and easier debugging of issues.

See [DESIGN.md](docs/DESIGN.md) for detailed design documentation.

## Contributing

Contributions are welcome! Please submit pull requests using
GitHub. Use [rbmk-project/issues](https://github.com/rbmk-project/issues)
to create issues and discuss features.

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```
