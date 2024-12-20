
# rbmk pipe read - Read from Named Pipe

## Usage

```
rbmk pipe read --writers N PIPE
```

## Description

Read from a named pipe in the current directory. Accepts exactly `N` writers
and multiplexes their output to stdout line by line. The command terminates
after all `N` writers have disconnected. If `N` is not specified
or is zero, this command will `exit 1` and print an error.

Each line from each writer is written atomically to stdout to prevent garbled
output when multiple writers are sending data simultaneously.

## Flags

### `--writers N`

Number of writers to expect (required). The command will accept exactly
this many connections before terminating.

## Examples

Read from two writers:

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

- Blocks until exactly `N` writers have connected and completed

- Each writer's output is handled line by line

- Lines from different writers may be interleaved

- Writers may disconnect early

- No new connections are accepted after `N` writers

## Exit Status

Returns `0` on success. Returns `1` on:

- Usage errors (missing pipe name or writers count)

- Connection errors

- I/O errors

## History

The `rbmk pipe read` command was introduced in RBMK v0.4.0.
