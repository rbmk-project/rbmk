# DNS data format

|              |                                                |
|--------------|------------------------------------------------|
| Author       | [@bassosimone](https://github.com/bassosimone) |
| Last-Updated | 2025-03-09                                     |

This document describes the format of DNS measurements emitted
by [rbmk](https://github.com/rbmk-project/rbmk) and implemented by
the [dnscore](https://github.com/rbmk-project/dnscore) library.


## Table of contents

- [Overview](#overview)
- [Address and Protocol](#address-and-protocol)
- [Query](#query)
- [Response](#response)
- [Example](#example)


## Overview

The [dnscore](https://github.com/rbmk-project/dnscore) library
emits a "query" event before sending a DNS query and one or more
"response" events when receiving responses. In general, we
expect a single "response" per query. However, when using DNS
over UDP, we may receive multiple responses for a single query,
due to network misconfiguration or internet censorship.

Specifically, the Great Firewall of China (GFW) may cause
multiple responses for the same query, with the last response
typically being the legitimate response from the server, and
previous responses being fake and for arbitrary addresses.


## Address and Protocol

A remote server is identified by an address and a protocol.

The protocol is an enumeration string with the following values:

- `"udp"` for DNS over UDP;

- `"tcp"` for DNS over TCP;

- `"dot"` for DNS over TLS;

- `"doh"` for DNS over HTTPS.

- `"doq"` for DNS over QUIC.

For `"udp"`, `"tcp"`, `"dot"`, and `"doq"` the address is a string
containing an IP address and a port separated by `":"`. When
the address is and IPv6 address, it is enclosed in square
brackets. The following are valid addresses:

- `"8.8.8.8:53"`;

- `"[2001:4860:4860::8888]:53"`.

For `"doh"`, the address is a string containing the URL of the
server. The following are valid addresses:

- `"https://dns.google/dns-query"`;

- `"https://cloudflare-dns.com/dns-query"`.


## Query

The JSON serialization of the query message contains
*at least* the following fields:

```JSON
{
  "msg":"dnsQuery",
  "dnsRawQuery":"",
  "serverAddr":"",
  "serverProtocol":"",
  "t":"",
  "protocol": ""
}
```

Where:

- `"msg"` (string) is the message type and is
always equal to `"dnsQuery"`;

- `"dnsRawQuery"` (base64 string) contains the
raw DNS query in base64 encoding;

- `"serverAddr"` (string) is the address of the
server, whose format depends on the protocol;

- `"serverProtocol"` (string) is the DNS protocol we
are using for the query (e.g., `"dot"`, `"doq"`);

- `"t"` (string) is the RFC3339 representation
of the time right before sending the query;

- `"protocol"` (string) is the network protocol we use (e.g., `"tcp"`, `"udp"`).

The current [dnscore](https://github.com/rbmk-project/dnscore)
implementation uses [log/slog](https://pkg.go.dev/log/slog), which
causes the generated message to contain additional fields that
you can safely ignore when processing the message.


## Response

The JSON serialization of the response message contains
*at least* the following fields:

```JSON
{
  "msg":"dnsResponse",
  "localAddr": "",
  "dnsRawQuery":"",
  "dnsRawResponse":"",
  "remoteAddr": "",
  "serverAddr":"",
  "serverProtocol":"",
  "t0":"",
  "t":"",
  "protocol": ""
}
```

Where:

- `"msg"` (string) is the message type and is
always equal to `"dnsQuery"`;

- `"localAddr"` (string) local address and port
of the socket we're using;

- `"dnsRawQuery"` (string) contains the
raw DNS query in base64 encoding;

- `"dnsRawResponse"` (string) contains the
raw DNS response in base64 encoding;

- `"remoteAddr"` (string) remote address and port
of the socket we're using;

- `"serverAddr"` (string) is the address of the
server, whose format depends on the protocol;

- `"serverProtocol"` (string) is the DNS protocol we
are using for the query (e.g., `"dot"`, `"doq"`);

- `"t0"` (string) is the RFC3339 representation of
the time right before sending the query;

- `"t"` (string) is the RFC3339 representation
of the time when the response was received;

- `"protocol"` (string) is the network protocol we use (e.g., `"tcp"`, `"udp"`).

The current [dnscore](https://github.com/rbmk-project/dnscore)
implementation uses [log/slog](https://pkg.go.dev/log/slog), which
causes the generated message to contain additional fields that
you can safely ignore when processing the message.


## Example

Here is an example of a `"dnsQuery"` message:

```JSON
{
  "msg":"dnsQuery",
  "dnsRawQuery":"yHUBAAABAAAAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABAAApBNAAAAAAAAA=",
  "serverAddr":"8.8.8.8:53",
  "serverProtocol":"udp",
  "t":"2024-11-18T15:31:53.05491+01:00",
  "protocol": "udp"
}
```

Here is an example of a `"dnsResponse"` message:

```JSON
{
  "msg":"dnsResponse",
  "localAddr": "130.192.91.211:32769",
  "dnsRawQuery":"yHUBAAABAAAAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABAAApBNAAAAAAAAA=",
  "dnsRawResponse":"yHWBgAABAAEAAAABA3d3dwdleGFtcGxlA2NvbQAAAQABwAwAAQABAAANVAAEXbjXDgAAKQIAAAAAAAAA",
  "remoteAddr": "8.8.8.8:53",
  "serverAddr":"8.8.8.8:53",
  "serverProtocol":"udp",
  "t0":"2024-11-18T15:31:53.05491+01:00",
  "t":"2024-11-18T15:31:53.072107+01:00",
  "protocol": "udp"
}
```
