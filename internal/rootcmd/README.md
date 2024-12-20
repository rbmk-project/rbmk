
# rbmk - Really Basic Measurement Kit

## Usage

```
rbmk COMMAND [args...]
```

RBMK (Really Basic Measurement Kit) is a command-line utility
to facilitate network exploration and measurements.

## Commands

### Core measurement commands

* `curl` - Measures HTTP/HTTPS endpoints with `curl(1)`-like syntax.
* `dig` - Performs DNS measurements with `dig(1)`-like syntax.
* `nc` - Measures TCP and TLS endpoints with an OpenBSD `nc(1)`-like syntax.
* `stun` - Performs STUN binding requests to discover public IP address.

### Unix-like Commands for Scripting

* `cat` - Concatenates files to standard output.
* `ipuniq` - Shuffle, deduplicate, and format IP addresses.
* `mkdir` - Creates directories.
* `mv` - Moves (renames) files and directories.
* `pipe` - Creates named pipes for inter-process communication.
* `rm` - Removes files and directories.
* `sh` - Runs POSIX shell scripts.
* `tar` - Creates tar archives.
* `timestamp` - Prints filesystem-friendly UTC timestamp.

### Plugins

* `plugin` - Manages RBMK plugins.

### Help Commands

* `intro` - Shows a brief introduction with usage examples.
* `tutorial` - Provides comprehensive usage documentation.

## Getting Started

New to RBMK? Try `rbmk intro` to get started!

Run `rbmk COMMAND --help` for more information about `COMMAND`.

## License

```
SPDX-License-Identifier: GPL-3.0-or-later
```

## Reporting Bugs

Please, use the [rbmk-project/issues](https://github.com/rbmk-project/issues)
repository to report bugs or suggest improvements.
