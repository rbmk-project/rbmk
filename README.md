# Really Basic Measurement Kit

[![GoDoc](https://pkg.go.dev/badge/github.com/rbmk-project/rbmk)](https://pkg.go.dev/github.com/rbmk-project/rbmk)

RBMK (Really Basic Measurement Kit) is a command-line utility
to facilitate network exploration and measurements. It provides
two atomic operations (`dig` and `curl`) that you can
compose together to perform modular network measurements where
you can observe each operation in isolation.

## Features

- Modular design with separate DNS and HTTP(S) measurements
- CLI-first approach with composable commands
- Extensive structured logging for detailed analysis
- Support for multiple DNS protocols (UDP, TCP, DoT, DoH)
- HTTP(S) measurements with granular control

- Core Measurement Commands:
  - `dig`: DNS measurements with multiple protocols
  - `curl`: HTTP(S) endpoint measurements

- Scripting Support:
  - Built-in POSIX shell interpreter
  - Cross-platform Unix-like commands
  - Script generation tools
  - Consistent workspace organization

The tool is designed to support both general use and measurement-specific
features, with careful consideration of concurrent operations and
extensive testing capabilities.

## Installation

```sh
go install github.com/rbmk-project/rbmk/cmd/rbmk@latest
```

## Quick Start

```sh
# Resolve a domain name
$ rbmk dig +short=ip example.com
93.184.215.14

# Make an HTTP request
$ rbmk curl https://example.com/

# Combine dig and curl for step-by-step measurement
$ IP=$(rbmk dig +short=ip example.com|head -n1)
$ rbmk curl --resolve example.com:443:$IP https://example.com/

# Collect measurement data in flat JSONL format
$ rbmk dig --logs dns.jsonl example.com
$ rbmk curl --logs http.jsonl https://example.com/
```

For a quick introduction with more examples:

```sh
$ rbmk intro
```

## Commands

Core Measurement Commands:
- `curl`: Measures HTTP/HTTPS endpoints with `curl(1)`-like syntax.
- `dig`: Performs DNS measurements with `dig(1)`-like syntax.

Unix-like Commands for Scripting:
- `cat`: Concatenates files.
- `ipuniq`: Filter out duplicate IP addresses.
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
