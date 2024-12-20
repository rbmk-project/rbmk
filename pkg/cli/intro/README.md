
# rbmk intro - Quick Start Guide

## Overview

RBMK is designed for composable network measurements. The two core commands
are `dig` for DNS queries and `curl` for HTTP(S) requests. You can combine
these to perform step-by-step measurements.

## Basic Examples

### DNS resolution

Get IP address for a domain

```
$ rbmk dig +short=ip example.com
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

### Combining Commands

Separate DNS and HTTP measurements:

```
$ IP=$(rbmk dig +short=ip example.com|head -n1)
$ rbmk curl --resolve example.com:443:$IP https://example.com/
```

### Collecting Structured Logs

Save measurement logs:

```
$ rbmk dig --logs dns.jsonl example.com
$ rbmk curl --logs http.jsonl https://example.com/
```

## Benefits

This modular approach helps isolate different aspects of network measurements
and makes it easier to understand where issues occur.

## Next Steps

* Run `rbmk dig --help` for detailed DNS measurement options

* Run `rbmk curl --help` for detailed HTTP measurement options

* Try `rbmk tutorial` for a comprehensive guide

## History

The `rbmk intro` command was introduced in RBMK v0.1.0.
