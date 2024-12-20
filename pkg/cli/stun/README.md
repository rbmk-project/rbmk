
# rbmk stun - STUN Measurements

## Usage

```
rbmk stun [flags] ENDPOINT
```

## Description

Send a STUN Binding Request to the given `ENDPOINT` and print the reflexive
transport address (public IP address and port) to standard output.

## Arguments

### `ENDPOINT`

The ENDPOINT argument should be in the form `HOST:PORT`. For example:

- `stun.l.google.com:19302`

- `74.125.250.129:19302`

- `[2001:4860:4864:5:8000::1]:19302`

We recommend using IPv4 and IPv6 addresses explicitly, to collect both
the externally observable IPv4 and IPv6 addresses.

## Flags

### `-h, --help`

Print this help message.

### `--logs FILE`

Writes structured logs to the given `FILE`. If `FILE` already exists, we
append to it. If `FILE` does not exist, we create it. If `FILE` is a single
dash (`-`), we write to the stdout.

### `--max-time DURATION`

Sets the maximum time that the STUN transaction operation is allowed to take
in seconds (e.g., `--max-time 5`). If this flag is not specified, the
default max time is 30 seconds.

### `--measure`

Do not exit with `1` if communication with the endpoint fails. Only exit
with `1` in case of usage errors, or failure to process inputs. You should
use this flag inside measurement scripts along with `set -e`. Errors are
still printed to stderr along with a note indicating that the command is
continuing due to this flag.

## Examples

Basic usage:

```
$ rbmk stun 74.125.250.129:19302
192.0.2.1:54321
```

Save structured logs to a file:

```
$ rbmk stun --logs stun.jsonl 74.125.250.129:19302
```

## Exit Status

Returns `0` on success. Returns `1` on:

- Usage errors (invalid flags, missing arguments, etc).

- File operation errors (cannot open/close files).

- Measurement failures (unless `--measure` is specified).

## History

The `rbmk stun` command was introduced in RBMK v0.3.0.
