
# rbmk stun - STUN Measurements

## Usage

```
rbmk stun [flags] ENDPOINT
```

## Description

Send a STUN Binding Request to the given ENDPOINT and print the reflexive
transport address (public IP address and port) to standard output.

## Arguments

### `ENDPOINT`

The ENDPOINT argument should be in the form HOST:PORT. For example:

- `stun.l.google.com:19302`

- `74.125.250.129:19302`

- `[2001:4860:4864:5:8000::1]:19302`

We recommend using IPv4 and IPv6 addresses explicitly, to collect both
the externally observable IPv4 and IPv6 addresses.

## Flags

### `-h, --help`

Print this help message.

### `--logs FILE`

Writes structured logs to the given FILE. If FILE already exists, we
append to it. If FILE does not exist, we create it. If FILE is a single
dash (`-`), we write to the stdout.

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

This command exits with `0` on success and `1` on failure.