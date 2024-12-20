
# rbmk curl - HTTP Measurements

## Usage

```
rbmk curl [flags] URL
```

## Description

A subset of `curl(1)` functionality focused on network measurements. We only
support measuring `http://` and `https://` URLs.

## Flags

### `-h, --help`

Print this help message.

### `--logs FILE`

Writes structured logs to the given FILE. If FILE already exists, we
append to it. If FILE does not exist, we create it. If FILE is a single
dash (`-`), we write to the stdout. If you specify `--logs` multiple
times, we write to the last FILE specified.

### `--max-time DURATION`

Sets the maximum time that the transfer operation is allowed to take
in seconds (e.g., `--max-time 5`). If this flag is not specified, the
default max time is 30 seconds.

### `--measure`

Do not exit with `1` if communication with the server fails. Only exit
with `1` in case of usage errors, or failure to process inputs. You should
use this flag inside measurement scripts along with `set -e`. Errors are
still printed to stderr along with a note indicating that the command is
continuing due to this flag.

### `-o, --output FILE`

Write the response body to FILE instead of using the stdout.

### `--resolve HOST:PORT:ADDR`

Use ADDR instead of DNS resolution for HOST:PORT.

Implementation note: we ignore the port and replace the HOST with
ADDR for every port number. Additionally, when using this flag, the
DNS lookup fails with "no such host" if the URL host is not HOST.

### `-v, --verbose`

Make the operation more talkative.

### `-X, --request METHOD`

Use the given request METHOD instead of GET.

## Examples

The following invocation prints the response body
of the `https://example.com/` website URL:

```
$ rbmk curl https://example.com/
```

To also print request and response headers, use `-v`:

```
$ rbmk curl -v https://example.com/
```

To save structured logs to `logfile.jsonl` use `--logs`:

```
$ rbmk curl --logs logfile.jsonl https://example.com/
```

To save the response body to `output.txt` use `-o`:

```
$ rbmk curl -o output.txt https://example.com/
```

To use a previously resolved IP address, use `--resolve`:

```
$ rbmk curl --resolve example.com:443:93.184.215.14 https://example.com/
```

## Exit Status

Returns `0` on success. Returns `1` on:

- Usage errors (invalid flags, missing arguments, etc).

- File operation errors (cannot open/close files).

- Measurement failures (unless `--measure` is specified).

## History

The `rbmk curl` command was introduced in RBMK v0.1.0.
