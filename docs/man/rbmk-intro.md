
# rbmk intro - Quick Start Guide

## Overview

RBMK is designed for composable network measurements. The two core commands
are `dig` for DNS queries and `curl` for HTTP(S) requests. You can combine
these to perform step-by-step measurements.

## Basic Examples

### DNS resolution

Get IP address for a domain

```
$ rbmk dig +short=ip example.com | rbmk head -n1
93.184.215.14
```

### HTTP Measurements

Fetch a webpage:

```
$ rbmk curl https://example.com/
```

### Using Specific IP address

Fetch webpage using a specific IP address:

```
$ rbmk curl --resolve example.com:443:93.184.215.14 https://example.com/
```

Note that `rbmk curl` ignores the port passed to `--resolve` and uses
the given address for all ports of the given domain.

### Combining Commands

Separate DNS and HTTP measurements:

```bash
addr=$(rbmk dig +short=ip example.com | rbmk head -n1)
rbmk curl --resolve example.com:443:$addr https://example.com/
```

Or, to measure all the available IP addresses:

```bash
for addr in $(rbmk dig +short=ip example.com); do
    rbmk curl --resolve example.com:443:$addr https://example.com/
done
```

### Collecting Structured Logs

Save measurement logs:

```
$ rbmk dig --logs dns.jsonl example.com
$ rbmk curl --logs http.jsonl https://example.com/
```

Use `--logs -` to emit the logs to the standard output.

## Benefits

This modular approach helps isolate different aspects of network measurements
and makes it easier to understand where issues occur.

## Next Steps

* Run `rbmk dig --help` for detailed DNS measurement options

* Run `rbmk curl --help` for detailed HTTP measurement options

* Try `rbmk tutorial` for a comprehensive guide

## History

The `rbmk intro` command was introduced in RBMK v0.1.0.
