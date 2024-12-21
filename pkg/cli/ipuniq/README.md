
# rbmk ipuniq - IP Address Processing

## Usage

```
rbmk ipuniq [flags] [FILE...]
```

## Description

Read IP addresses from files or stdin and remove duplicates. We expect
input files, or stdin, to contain one IP address per line. Use `-` as the
file name to explicitly indicate you want to read from the stdin.

More specifically, `rbmk ipuniq`:

- Reads IPv4 and IPv6 addresses from input files or the stdin

- Normalizes different textual representations

- Removes duplicates

- Optionally, randomly shuffles the resulting addresses

You enable random shuffling with `--random`. When not using `--random`, we
stream unique addresses as soon as they are read. The streaming functionality
was implemented in RBMK v0.4.0.

The `-E, --from-endpoints` flag modifies the command behaviour to assume the
input contains endpoints rather than just IP addresses (see below).

Note that any input line that does not contain a valid IP address or
endpoint is skipped and a warning is emitted.

## Flags

### `-E, --from-endpoints`

Assume that input already contains endpoints, that is `addr:port`, where
the `addr` is an IP address, and there is `[` and `]` around IPv6
addresses. Strip the port, and just retain the IP address for processing
and emitting according to other `rbmk ipuniq` flags.

### `-f, --fail`

Cause the tool to fail (rather than emitting a warning) if an input
line does not contain a valid IP address or endpoint (if `-E`).

### `-h, --help`

Print this help message.

### `-p, --port PORT`

Format output as `ADDRESS:PORT` endpoints, adding [] brackets for IPv6
addresses as needed. This flag can be specified multiple times
to generate endpoints for multiple ports (e.g., `-p 80 -p 443 -p 22`
generates HTTP, HTTPS, and SSH endpoints). When no ports are
specified, we output IP addresses without ports. Each `PORT` must
be a valid port number (0-65535).

### `--only ipv4|ipv6`

Only output addresses belonging to the specific IP version.

This flag has been introduced in RBMK v0.11.0.

### `-r, --random`

Buffers and randomly shuffles the addresses before output. This
flag has been introduced in v0.4.0.

## Examples

### Process DNS Resolution Results

Collect and measure unique IP addresses:

```
$ rbmk dig +short=ip example.com A > dig_A.txt

$ rbmk dig + short=ip example.com AAAA > dig_AAAA.txt

$ for ipAddr in $(rbmk ipuniq dig_A.txt dig_AAAA.txt); do \
  rbmk curl --resolve "example.com:443:${ipAddr}" https://example.com/ \
done
```

Randomize IP addresses read from stdin:

```
$ rbmk ipuniq --random
```

Filters stdin and immediately emits unique addresses:

```
$ rbmk ipuniq
```

### Generate STUN Endpoints

Create endpoints for STUN measurements:

```
$ rbmk dig +short=ip stun.l.google.com A > ips.txt

$ rbmk dig +short=ip stun.l.google.com AAAA >> ips.txt

$ rbmk ipuniq --port 19302 stun_A.txt stun_AAAA.txt
```

### Generate Multiple Endpoints

Create multiple endpoints for each IP addr:

```
$ rbmk ipuniq --port 80 --port 443 ips.txt
```

### Filtering a list of endpoints

Filter a list the endpoints and just keep IP addresses:

```
$ echo -e '10.0.0.1:80\n10.0.0.1:443\n127.0.0.1:111' | rbmk ipuniq -E
10.0.0.1
127.0.0.1
```

Same but using IPv6 endpoints:

```
$ echo -e '[::1]:80\n[::1]:443' | rbmk ipuniq -E
::1
```

### Exit Status

This command exits with `0` on success and `1` on failure.

## Bugs

When running a command such as:

```
$ rbmk ipuniq
```

we keep the `stdin` in line-oriented mode, which means that you
can edit the input before pressing enter. However, this also implies
that `^C` does not interrupt reading from the `stdin`, because
the terminal driver is blocked reading until the EOL. The symptom
of this would be:

```
$ rbmk ipuniq
^C
```

where the program does not exit. To exit, insert an explicit
EOL character (e.g., `^D` on Unix and `^Z` + `Return` on Windows).

### History

The `--only` flag was introduced in RBMK v0.11.0.

Before RBMK v0.4.0, this command always randomly shuffled the
addresses. Afterwards, one must use `--random` explicitly.

This command was introduced in RBMK v0.2.0.
