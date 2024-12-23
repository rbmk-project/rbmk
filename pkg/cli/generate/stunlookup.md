
# rbmk generate stun_lookup - STUN Lookup Script Generation

## Usage

```
rbmk generate stun_lookup [flags]
```

## Description

Generate a shell script that performs STUN lookups using multiple STUN
servers to discover the public IP address(es) of the current host.

## Flags

### `-h, --help`

Print this help message.

### `--input ENDPOINT`

Add STUN endpoint(s) to measure. Each endpoint must be in the form of
`HOST:PORT`. Can be specified multiple times. If not specified, uses
a default list of well-known STUN servers. If `HOST` is an IPv6 address,
it must be enclosed in square brackets. For example, the following:

- `stun.l.google.com:19302`
- `[2001:4860:4864:5:8000::1]:19302`
- `74.125.250.129:19302`

are all valid STUN endpoints.

### `--input-file FILE`

Read STUN endpoints from file(s). Each line should contain one endpoint
in the form `HOST:PORT`. Can be specified multiple times.

### `--minify`

Minify the output script by removing comments and unnecessary whitespace.

### `--output FILE`

Write script to FILE instead of stdout. Use `-` for stdout.

## Examples

Generate script with default STUN servers:

```
$ rbmk generate stun_lookup > script.sh
```

Use specific STUN servers:

```
$ rbmk generate stun_lookup \
    --input stun.l.google.com:19302 \
    --input stun.ekiga.net:3478 > script.sh
```

Read endpoints from file:

```
$ rbmk generate stun_lookup --input-file servers.txt > script.sh
```

## Generated Script

The generated script:

1. Performs STUN lookups using specified servers
2. Collects measurement data in JSON format
3. Shows progress with a progress bar
4. Archives results in a tarball
5. Accepts command-line arguments to customize behavior

See the script's embedded documentation for details.

## History

The `rbmk generate stun_lookup` command was introduced in RBMK v0.12.0.
