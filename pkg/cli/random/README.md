
# rbmk random - Random Bytes Generation

## Usage

```
rbmk random [flags]
```

## Description

Generate random bytes using a cryptographically secure random number
generator and print them as hexadecimal to stdout.

## Flags

### `-h, --help`

Print this help message.

### `--bytes COUNT`

Number of random bytes to generate. The default is 4 bytes.

## Examples

Generate 4 random bytes (default):

```
$ rbmk random
a1b2c3d4
```

Generate 16 random bytes:

```
$ rbmk random --bytes 16
a1b2c3d4e5f6789012345678deadbeef
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

The `rbmk random` command was introduced in RBMK v0.12.0.
