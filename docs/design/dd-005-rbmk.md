# RBMK Design Document

|              |                                                |
|--------------|------------------------------------------------|
| Author       | [@bassosimone](https://github.com/bassosimone) |
| Last-Updated | 2024-12-09                                     |

This document describes the design of the [rbmk](
https://github.com/rbmk-project/rbmk) repository, which
contains the `rbmk` command implementation.


## Overview

The `rbmk` command is a simple command line tool used
to perform network measurements. We implement three toplevel
core commands:

- `rbmk dig`, emulating a subset of `dig(1)` to perform
DNS network measurements;

-  `rbmk curl`, emulating a subset of `curl(1)` to
perform HTTP/HTTPS network measurements;

- `rbmk nc`, emulating a subset of `nc(1)` to perform
TCP and TLS endpoint measurements;

- `rbmk stun`, to resolve the public IP addresses.


## Core Commands

The general syntax of the `rbmk` command is:

```
rbmk [command] [flags] [arguments]
```

Flags will always follow the `command`. We will generally
try to use GNU-style flags unless we need other kind of flags
to emulate specific commands (e.g., `dig`).

In general, command line arguments will always take precedence over
environment variables, which in turn take precedence over default settings.

All commands will implement these help facilities:

- `rbmk [command] --help`

- `rbmk [command] -h`

- `rbmk help [command]`

These invocations will all print a detailed usage message.

Additionally, `rbmk` invoked without arguments will also print
its own help like it was invoked as `rbmk --help`.

The programm will not report a command line error if it
encounters the `--help` or `-h` flag, even in presence of
otherwise wrong flags or arguments.

The inline help will also mention the relevant environment
variables that influence the behavior of each command.

### Error Handling and Logging

Commands will follow these principles for errors:

- Human-readable errors to stderr.

- Structured error logging when `--logs` is specified.

- Exit 0 on success, exit 1 on failure.

- Consistent error formatting across commands.

The `--logs -` flag will write structured logs to stdout.

The `--measure` flag will prevent measurement errors (e.g., I/O
timeouts) from causing the program to exit with a non-zero status.

### rbmk dig

*Purpose*: perform and measure DNS lookups.

*Usage*: `rbmk dig [flags] [@SERVER] DOMAIN [TYPE] [options]`

The command line syntax is similar to the `dig` command, with some
RBMK specific additions, such as `--logs FILE` to specify where
to the write structured logs. (For more information on the structured
logs, see the [data-format](../../spec/data-format) spec.)

*Output*: the command primary output will mimic `dig` output
quite closely, without pedentically being a clone.

*Key Features*:

- Support for DNS over UDP, TCP, TLS, HTTPS.

- Mimic dig output and behavior.

- Possibility to collect structured logs.

*Invocation Example*:

```console
rbmk dig --logs dig.jsonl www.example.com A +short
```

*Structured Logs Example*:

```JavaScript
// ...
{
  "msg":"dnsQuery",
  "dnsRawQuery":"upwBAAABAAAAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABAAApBNAAAAAAAAA=",
  "serverAddr":"8.8.8.8:53",
  "serverProtocol":"udp",
  "t":"2024-11-30T22:13:42.442704+01:00",
  "protocol":"udp"
}
// ...
{
  "msg":"dnsResponse",
  "localAddr":"192.0.2.1:52956",
  "dnsRawQuery":"upwBAAABAAAAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABAAApBNAAAAAAAAA=",
  "dnsRawResponse":"upyBgAABAAEAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABwAwAAQABAAADhwAEXbjXDgAAKQIAAAAAAAAA",
  "remoteAddr":"8.8.8.8:53",
  "serverAddr":"8.8.8.8:53",
  "serverProtocol":"udp",
  "t0":"2024-11-30T22:13:42.442704+01:00",
  "t":"2024-11-30T22:13:42.459874+01:00",
  "protocol":"udp"
}
```

### rbmk curl

*Purpose*: measure HTTP/HTTPS endpoints.

*Usage*: `rbmk curl [flags] URL`

The command line syntax is similar to the `curl` command, with some
RBMK specific additions, such as `--logs FILE` to specify where
to the write structured logs.

*Output*: the command primary output will mimic `curl` output
quite closely, without pedentically being a clone.

*Key Features*:

- Support for HTTP and HTTPS.

- Support specifying an IP address via `--resolve` (a flag
that is also implemented by `curl` and allows for composing
`rbmk dig` and `rbmk curl` from a Unix shell).

- Mimic curl output and behavior.

- Possibility to collect structured logs.

*Invocation Example*:

```console
rbmk curl --logs curl.jsonl https://www.example.com/
```

*Structured Logs Example*:

```JavaScript
// ...
{
  "msg":"httpRoundTripDone",
  "httpMethod":"GET",
  "httpUrl":"https://www.example.com/",
  "httpRequestHeaders":{},
  "httpResponseStatusCode":200,
  "httpResponseHeaders":{
    "Accept-Ranges":["bytes"],
    "Age":["182207"],
    "Cache-Control":["max-age=604800"],
    "Content-Type":["text/html; charset=UTF-8"],
    "Date":["Sat, 30 Nov 2024 21:19:37 GMT"],
    "Etag":["\"3147526947+gzip\""],
    "Expires":["Sat, 07 Dec 2024 21:19:37 GMT"],
    "Last-Modified":["Thu, 17 Oct 2019 07:18:26 GMT"],
    "Server":["ECAcc (dcd/7D56)"],
    "Vary":["Accept-Encoding"],
    "X-Cache":["HIT"]
  },
  "localAddr":"[2001:db8::1]:59187",
  "protocol":"tcp",
  "remoteAddr":"[2606:2800:21f:cb07:6820:80da:af6b:8b2c]:443",
  "t":"2024-11-30T22:19:36.619372+01:00"
}
// ...
```

### rbmk nc

*Purpose*: measure TCP/TLS endpoints.

*Usage*: `rbmk nc [flags] HOST PORT`

The command line syntax is similar to the OpenBSD `nc(1)` command,
with some RBMK specific additions, such as `--logs FILE` to specify
where to the write structured logs.

*Key Features*:

- Support for TCP and TLS.

- Possibility to collect structured logs.

*Invocation Example*:

```console
rbmk nc --logs curl.jsonl -z 93.184.215.14 80
```

### rbmk stun

*Purpose*: resolve the public IP addresses.

*Usage*: `rbmk stun [flags] ENDPOINT`

The command syntax is not similar to any existing command and
is unique and specific to the `rbmk` tool.

*Output*: the command primary output consists of the address
of the public UDP endpoint that contacted the STUN server, thus
showing the public IP address of the machine.

*Key Features*:

- Possibility to collect structured logs.

*Invocation Example*:

```console
rbmk stun --logs - 74.125.250.129:19302
```

*Structured Logs Example*:

```JavaScript
// ...
{
  "msg":"stunReflexiveAddress",
  "stunReflexiveIPAddr":"192.0.2.1",
  "stunReflexivePort":64986
}
// ...
```

## Command Composition

It should be possible to compose `rbmk dig`, `rbmk curl`, `rbmk nc` and
`rbmk stun` together as illustrated by the following session example:

```console
% rbmk dig +short --logs dig.jsonl www.example.com A
93.184.215.14

% rbmk curl --logs curl.jsonl \
    --resolve www.example.com:443:93.184.215.14 \
    https://www.example.com/

# ... response body omitted
```

The `rbmk dig +short` command prints the resolved IP
addresses and `--resolve` allows to focus on a specific
endpoint to measure, thus implementing "step by step"
measurements (i.e., measurements where each operation that
may fail is executed independently).

See [dd-006-rbmk-scripting](dd-006-rbmk-scripting.md) for
more information about the scripting extensions.

## Implementation Approach

We want to start with a small and simple implementation, and
use scripts (either bash scripts or python scripts) to iterate
quickly and understand which additional functionality to add.

While the `rbmk` code derives from OONI Probe code to some
extent, the focus of this tool is not necessarily measuring
internet censorship, rather to more generically diagnose issues
and help measuring the performance of the network.

The code inside the `rbmk` repository should expose its building
blocks in reusable fashion via public packages, to support creating
more complex tools using `rbmk` as a library. Namely:

- `pkg/cli/dig` will expose the `main` of the `rbmk dig` command
and the specific "task" that executes the command itself.

- `pkg/cli/curl` will expose the `main` of the `rbmk curl` command
and the specific "task" that executes the command itself.

- `pkg/cli/nc` will expose the `main` of the `rbmk nc` command
and the specific "task" that executes the command itself.

- `pkg/cli/stun` will expose the `main` of the `rbmk stun` command
and the specific "task" that executes the command itself.

A "task" is a structure named `Task` with public fields that one
can set, which implements a `Run` method that executes the specified
command and return an error (or `nil` on success). For example:

```Go
type Task struct {
  URL        string    // Target URL
  LogsWriter io.Writer // Where to write structured logs
  Resolve    []string  // --resolve entries
}

func (t *Task) Run(/* ... */) error {
  // Implementation
}
```

The `main` is a public main-like function that parses command
line arguments, fills the `Task` and calls `Run` on it.

Additionally, we want robust integration testing where we use
network simulation (through, for example, the `netsim` package
inside the [x](dd-002-x.md) package collection) to provoke
specific network conditions and verify the emitted structured
logs are correct for the given condition.

## Design Principles

In terms of code structure, the overall aim is that of having
a fixed core set of functionality glued together by either
scripting code or by code (perhaps even written in Go) that
behaves like a sort of a script and glues thing together.

It is important to avoid cramming more complex feature together
into the core functionality, to allow the system to evolve
freely and avoid too much interlocking between the components.

More specifically, we consider core:

- The DNS lookup functionality.

- The HTTP/HTTPS endpoint measurement functionality.

- The `nc` functionality.

- The STUN client functionality.

- Emitting structured logs.

- Any package necessary to emulate real clients including
code for parroting TLS handshakes and code to simulate the
ordering of headers emitted by web browsers.

- The `rbmk` command itself.

Instead, we consider glue the set of scripts (or Go code)
that compose these functionality together to implement
specific network measurement algorithms.

## Versioning

We will follow semantic versioning. Because this is a research
and experimentation tool, and because Go complicates version
management for v1 and above, we will most likely not ever cut
a stable release and remain in the v0.x territory. That said, we
will take care to avoid breaking existing public code unless it
is needed to fix bugs or implement new important features.

## Documentation

The command line itself should be very informative and produce
output similar to the output produced by `go help`. The command
line itself should include extra documentation in form of
introductions and tutorials.

The above described `--help` and `-h` flags will help users to
immediately get comprehensive inline help.

The `rbmk intro` command will print a short introduction to
using the tool, while the `rbmk tutorial` command will print
a more detailed tutorial explaining how to measure more in
detail.

Beyond the `rbmk` command itself, we also aim to produce
comprehensive documentation as part of the
`rbmk-project.github.io` repository and website.
