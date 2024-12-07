
# rbmk pipe - Named Pipe Operations

## Usage

```
rbmk pipe COMMAND [args...]
```

## Description

Create and use named pipes for inter-process communication within
measurement scripts. Uses Unix domain sockets on both Unix systems
and modern Windows (10.0.17063+).

## Commands

### read

Read from a named pipe. Blocks until writer connects.

### write

Write to a named pipe. Blocks until reader is ready.

## Examples

Read from a pipe:

```
$ rbmk pipe read mypipe
```

Write to a pipe:

```
$ echo "data" | rbmk pipe write mypipe
```

Typical measurement pattern:

```bash
#!/bin/bash

# Write addresses as they become available
( ./rbmk dig +short=ip example.com A | ./rbmk pipe write addresses ) &
( ./rbmk dig +short=ip example.com AAAA | ./rbmk pipe write addresses ) &

# Use addresses as they become available
./rbmk pipe read --writers 2 addresses | while read addr; do
  ./rbmk curl --resolve example.com:443:$addr "https://example.com/"
done
```

## Notes

- Pipes are created relative to the current directory

- Pipes are automatically cleaned up when both ends close

- Maximum connection timeout is 1s

## Exit Status

Returns `0` on success. Returns `1` on:

- Usage errors

- Connection timeouts

- I/O errors
