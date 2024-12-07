
# rbmk ipuniq - IP Address Processing

## Usage

```
rbmk ipuniq [flags] FILE...
```

## Description

Read IP addresses from files, sort them, and remove duplicates. We expect
input files to contain one IP address per line.

More specifically, `rbmk ipuniq`:

- Reads IPv4 and IPv6 addresses from input files

- Normalizes different textual representations

- Removes duplicates

- Randomly shuffles the resulting addresses

## Flags

### `-h, --help`

Print this help message.

### `-p, --port PORT`

Format output as HOST:PORT endpoints, adding [] brackets for IPv6
addresses as needed. This flag can be specified multiple times
to generate endpoints for multiple ports (e.g., `-p 80 -p 443 -p 22`
generates HTTP, HTTPS, and SSH endpoints). When no ports are
specified, we output IP addresses without ports. Each PORT must
be a valid port number (0-65535).

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

### Exit Status

This command exits with `0` on success and `1` on failure.
