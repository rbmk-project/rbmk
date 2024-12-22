
# rbmk pipe - Named Pipe Operations

## Usage

```
rbmk pipe COMMAND [args...]
```

## Description

Create and use named pipes for inter-process communication within
measurement scripts. Uses Unix domain sockets on both Unix systems
and modern Windows systems (10.0.17063+).

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

## Bugs

Unix domain sockets have a platform-specific maximum path length
ranging from ~90 to ~108 bytes. If the path length exceeds the
maximum path length, you will see errors such as:

```
rbmk pipe read: cannot create pipe: listen unix .../pipe: bind: invalid argument
```

where `.../pipe` is a path that exceeds the maximum path length.

This limitation could interact with how `rbmk sh` executes `rbmk
COMMAND` commands (i.e., in the same process) as documented in
detail by the `rbmk sh --help` output.

To mitigate this issue, use paths relative to the current working
directory in your scripts and attemp to keep them short. Specifically,
you can create a unique directory for measuring with a short name,
using `rbmk timestamp --full` to generate a timestamp-based name with
nanosecond precision and possibly combining it with `rbmk random` to
add additional entropy. Then, once the measurement is complete, you
can move the results to a longer, more logical path name.

## History

The `rbmk pipe` command was introduced in RBMK v0.4.0.
