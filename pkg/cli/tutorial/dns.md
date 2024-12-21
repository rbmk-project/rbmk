
# DNS Tutorial

When measuring the DNS, often times your objective is to
obtain the following information:

1. *structured logs* of DNS queries and responses as well as
of the I/O events that occurred during the measurement;

2. *resolved IP addresses*.

Additionally, you may want to override the default server
used by `rbmk dig`, which is `8.8.8.8`, or the default
protocol, which is DNS over UDP.

We will start discussing common patterns for collecting the
typically required information, discuss how to manipulate
the resolved IP addresses, and then show how to change the
DNS server and protocol used by `rbmk dig`.

## Common Patterns

This section describes common usage patterns for DNS
measurements with increasing complexity.

### Basic Usage

To collect all the IP addresses associated with a domain name,
you need to issue two queries, one for `A` records (i.e., IPv4
addresses) and one for `AAAA` records (i.e., IPv6 addresses).

So, in general, you need two commands like this:

```bash
# Resolve IPv4
rbmk dig www.example.com A

# Resolve IPv6
rbmk dig www.example.com AAAA
```

### Saving IP Addresses to a File

The simplest pattern to obtain the IP addresses is to save
them to a file, one IP address per line. To this end, you can
use the `+short=ip` flag.

```bash
# Resolve IPv4
rbmk dig +short=ip www.example.com A > addrs4.txt

# Resolve IPv6
rbmk dig +short=ip www.example.com AAAA > addrs6.txt
```

### Ignoring Measurement Errors

Also, when measuring, you do not want measurement errors
to cause the program to `exit 1`, which would cause a script
using `set -e` to terminate immediately. To this end, use
the `--measure` flag, which causes `rbmk dig` to only `exit 1`
on usage errors (e.g., invalid command line flags):

```bash
# Ensure that usage errors terminate the script, that
# undefined variables cause errors, and that any command
# that fails in a pipeline causes an error.
set -euo pipefail

# Resolve IPv4
rbmk dig --measure +short=ip www.example.com A > addrs4.txt

# Resolve IPv6
rbmk dig --measure +short=ip www.example.com AAAA > addrs6.txt
```

### Collecting Structured Logs

Additionally, it is most often useful to collect structured
logs, for diagnostic and measurement purposes, which you
can do using the `--logs FILE` flag:

```bash
# Ensure that usage errors terminate the script, that
# undefined variables cause errors, and that any command
# that fails in a pipeline causes an error.
set -euo pipefail

# Resolve IPv4
rbmk dig --measure \
    +short=ip \
    --logs dig4.jsonl \
    www.example.com A > addrs4.txt

# Resolve IPv6
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    www.example.com AAAA > addrs6.txt
```

### Collecting Stderr Messages

Moreover, it may be useful to collect the messages emitted
to the standard error, if any, which you can do using the
`2>` operator as follows:

```bash
# Ensure that usage errors terminate the script, that
# undefined variables cause errors, and that any command
# that fails in a pipeline causes an error.
set -euo pipefail

# Resolve IPv4
rbmk dig --measure \
    +short=ip \
    --logs dig4.jsonl \
    www.example.com A \
    1> addrs4.txt \
    2> dig4.err

# Resolve IPv6
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

### Running Parallel Lookups

The `dig` command uses a *default timeout* of 5 seconds. You may
still want to run lookups in parallel, to speed-up things in case
there are timeouts. You can do this as follows:

```bash
# Ensure that usage errors terminate the script, that
# undefined variables cause errors, and that any command
# that fails in a pipeline causes an error.
set -euo pipefail

# Resolve IPv4 in the background
rbmk dig --measure \
    +short=ip \
    --logs dig4.jsonl \
    www.example.com A \
    1> addrs4.txt \
    2> dig4.err &

# Resolve IPv6 in the background
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err &

# Wait for both to finish
wait
```

### Waiting for Duplicate Responses

Some censored networks cause duplicate responses to queries
where subsequent responses are not identical. This is the case,
for example, of the Great Firewall of China (GFW). Note that
this behaviour is only possible when using DNS-over-UDP, which
is the default protocol. While `rbmk dig` does not show such
duplicate responses by default, you can force waiting for
such duplicates using the `+udp=wait-duplicates` query option:

```bash
# Ensure that usage errors terminate the script, that
# undefined variables cause errors, and that any command
# that fails in a pipeline causes an error.
set -euo pipefail

# Resolve IPv4 in the background
rbmk dig --measure \
    +short=ip \
    +udp=wait-duplicates \
    --logs dig4.jsonl \
    www.example.com A \
    1> addrs4.txt \
    2> dig4.err &

# Resolve IPv6 in the background
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +udp=wait-duplicates \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err &

# Wait for both to finish
wait
```

## Manipulating Resolved IP Addresses

The simplest approach is that of collating the resolved
IPv4 and IPv6 addresses using `rbmk cat`:

```bash
addrs=$(rbmk cat addrs4.txt addrs6.txt)
```

However, you may want to ensure that IP addresses
are unique, using `rbmk ipuniq`:

```bash
addrs=$(rbmk cat addrs4.txt addrs6.txt | rbmk ipuniq)
```

If you need to *transform* the IP addresses to endpoints, for
example by appending a port number, you can use the `--port` flag:

```bash
epnts=$(rbmk cat addrs4.txt addrs6.txt | rbmk ipuniq --port 80)
```

You can also randomize the IP addresses order with the `--random` flag:

```bash
epnts=$(rbmk cat addrs4.txt addrs6.txt | rbmk ipuniq --random --port 80)
```

Additionally, you can select only IPv4 addresses using `--only ipv4`
and only IPv6 addresses using `--only ipv6`:

```bash
# Generate IPv4 endpoints
epnts4=$(rbmk cat all.txt | rbmk ipuniq --only ipv4 --port 80)

# Generate IPv6 endpoints
epnts4=$(rbmk cat all.txt | rbmk ipuniq --only ipv6 --port 80)
```

## Changing the DNS Server and Protocol

Previous `rbmk dig` invocations were omitting the DNS server
and protocol to use in most cases. For example:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

is equivalent to:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +udp \
    @8.8.8.8 \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

where the `+udp` protocol selector and the `@8.8.8.8` server
selector have been explicitly specified.

### Changing the Protocol

You have previously seen how to use the `+udp=wait-duplicates`
selector to force waiting for duplicates while using DNS-over-UDP.

Additionally, you can force using DNS-over-TCP with`+tcp`:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +tcp \
    @8.8.8.8 \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

Likewise, you can force using DNS-over-TLS with `+tls`:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +tls \
    @8.8.8.8 \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

And, you can use DNS-over-HTTPS with `+https`:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +https \
    @8.8.8.8 \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

### Changing the Server

To change the server, modify the IP addresss after the `@` symbol.

For example:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +https \
    @8.8.4.4 \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

uses `8.8.4.4` instead of `8.8.8.8`.

You *may* specify a domain name after the `@` symbol, rather
than an IP address. In such a case, the code will resolve the
given domain name using the system resolver, and then try
all the available IP addresses in sequence, until one works.

For example:

```bash
rbmk dig --measure \
    --logs dig6.jsonl \
    +short=ip \
    +https \
    @dns.google \
    www.example.com AAAA \
    1> addrs6.txt \
    2> dig6.err
```

For measuring, we *recommend* using IP addresses directly,
which allows to control which IP address is being used.

## Next Steps

- Try running all the above commands and inspect the
generated structured logs.
