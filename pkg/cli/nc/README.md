# rbmk nc - TCP/TLS Client

## Usage

```
rbmk nc [flags] HOST PORT
```

## Description

The `rbmk nc` command emulates a subset of the OpenBSD `nc(1)` command,
including connecting to remote TCP/TLS endpoints, scanning for open ports,
sending and receiving data over the network.

The `HOST` may be a domain name, an IPv4 address, or an IPv6 address. When
using a domain name, we use the system resolver to resolve the name to a
list of IP addresses and try all of them until one succeeds. For measuring,
it is recommended to specify an IP address directly.

## Flags

### `--alpn PROTO`

Specify ALPN protocol(s) for TLS connections. Can be specified
multiple times to support protocol negotiation. For example:

```
--alpn h2 --alpn http/1.1
```

Must be used alongside the `--tls` flag.

### `-c, --tls`

Perform a TLS handshake after a successful TCP connection.

### `-h, --help`

Print this help message.

### `--logs FILE`

Writes structured logs to the given `FILE`. If `FILE` already exists, we
append to it. If `FILE` does not exist, we create it. If `FILE` is a single
dash (`-`), we write to the stdout. If you specify `--logs` multiple
times, we write to the last `FILE` specified.

### `--measure`

Do not exit with `1` if communication with the server fails. Only exit
with `1` in case of usage errors, or failure to process inputs. You should
use this flag inside measurement scripts along with `set -e`. Errors are
still printed to stderr along with a note indicating that the command is
continuing due to this flag.

### `--sni SERVER_NAME`

Specify the server name for the SNI extension in the TLS
handshake. For example:

```
--sni www.example.com
```

Must be used alongside the `--tls` flag.

### `-T, --option OPTION`

Set a specific `OPTION`. Can be specified multiple times to provide
multiple options. We currently accept the following options:

- `noverify` (added in RBMK v0.11.0): Do not verify the server's
certificate, which is useful for measuring SNI-based blocking using
arbitrary TLS endpoints as test helpers. This option only takes
effect when combined with the `-c, --tls` flag.

The `-T` flag was introduced in RBMK v0.11.0.

### `-v`

Print more verbose output.

### `-w, --timeout TIMEOUT`

Time-out I/O operations (connect, recv, send) after
a `TIMEOUT` number of seconds.

### `-z, --scan`

Without `--tls`, perform a port scan and report whether the
remote port is open. With `--tls`, perform a TLS handshake
and then close the remote connection.

## Examples

Basic TCP connection to HTTP port:

```
$ rbmk nc example.com 80
```

TLS connection with HTTP/2 and HTTP/1.1 ALPN:

```
$ rbmk nc -c --alpn h2 --alpn http/1.1 example.com 443
```

Check if port is open (scan mode) with a five seconds timeout:

```
$ rbmk nc -z -w5 example.com 80
```

Same as above but also perform a TLS handshake:

```
$ rbmk nc --alpn h2 --alpn http/1.1 -z -c -w5 example.com 443
```

Checking for SNI based blocking:

```
$ rbmk nc -zcw5 --alpn h2 --alpn http/1.1 -T noverify \
    --sni x.com example.com 443
```

Saving structured logs:

```
$ rbmk nc --logs conn.jsonl example.com 80
```

## Exit Status

The nc utility exits with `0` on success and `1` on error.

## Bugs

When running a command such as:

```
$ rbmk nc example.com 80
```

we keep the `stdin` in line-oriented mode, which allows to send
protocol data on a line-by-line basis. However, this also implies
that `^C` does not interrupt reading from the `stdin`, because
the terminal driver is blocked reading until the EOL. The symptom
of this would be:

```
$ rbmk nc example.com 80
^C
```

where the program does not exit. To exit, insert an explicit
EOL character (e.g., `^D` on Unix and `^Z` + `Return` on Windows).

## History

The `rbmk nc` command was introduced in RBMK v0.6.0.
