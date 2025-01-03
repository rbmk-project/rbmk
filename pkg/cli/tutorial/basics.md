
# RBMK Basics Tutorial

This tutorial introduces the fundamental concepts and usage patterns
of the RBMK measurement toolkit.


## Core Concepts

RBMK is designed for composable network measurements. Each command
performs a specific measurement task and can be combined with others
to address complex measurement scenarios.

All commands accept the `-h` or `--help` flag to display detailed
usage information. For example, `rbmk dig --help` displays the help
text associated with the `rbmk dig` subcommand.

All commands exit with `0` on success and with `1` on failure. Note
that failure includes both usage errors (e.g., passing an invalid flag
to a command) and measurement errors (e.g., a network timeout). The
`--measure` flag disables the exit-on-failure behavior for measurement
errors, thus allowing shell scripts to use `set -e`.


## Basic Commands

1. DNS Resolution (`rbmk dig`)

Resolve domain names and collect DNS measurement data:

```
$ rbmk dig +short=ip example.com
```

Run `rbmk dig --help` for additional help.

2. HTTP Measurements (`rbmk curl`)

Measure HTTP/HTTPS endpoints:

```
$ rbmk curl https://example.com/
```

Run `rbmk curl --help` for additional help.

3. STUN Probing (`rbmk stun`)

Discover your public IP address:

```
$ rbmk stun stun.l.google.com:19302
```

Run `rbmk stun --help` for additional help.

4. Netcat (`rbmk nc`)

Check whether a port is open:

```
$ rbmk nc -zv example.com 80
```

Same as above, but also checking for TLS reachability:

```
$ rbmk nc --alpn h2 --alpn http/1.1 -zvc example.com 443
```

Run `rbmk nc --help` for additional help.


## Combining Commands

Commands can be combined to perform detailed measurements:

```bash
addr=$(rbmk dig +short=ip example.com | rbmk head -n1)
rbmk curl --resolve "example.com:443:$addr" https://example.com/
```

And:

```bash
for addr in $(rbmk dig +short=ip example.com); do
    rbmk curl --resolve example.com:443:$addr https://example.com/
done
```

Combining measurements allows to isolate network operations and
analyze their failure in isolation. Additionally, by combining
operations, we can select which IP addresses to measure for a given
domain name, which allows us to investigate whether all the available
addresses for a domain name are reachable and working as intended.


## Structured Logging

Use `--logs` to collect detailed measurement data:

```sh
# Saves structured logs to the dns.jsonl file
rbmk dig --logs dns.jsonl example.com

# Saves structured logs to the http.jsonl file
rbmk curl --logs http.jsonl https://example.com/
```

The measurement data consists of a sequence of lines in JSON format
(also known as JSONL format). The data format emitted by commands
is documented in the [RBMK data format specification].

[RBMK data format specification]: https://github.com/rbmk-project/rbmk-project.github.io/tree/main/docs/spec/data-format

Using `--logs -` causes the command to emit logs to the standard output.


## Next Steps

- Try `rbmk tutorial dns` for DNS measurement patterns.

- Run `rbmk dig --logs - +noall example.com | jq` to see the structured logs.

- Try `rbmk tutorial http` for HTTP measurement patterns.

- Run `rbmk curl --logs - https://example.com/ | jq` to see the structured logs.

- Use `rbmk COMMAND --help` for detailed documentation on a `COMMAND`.
