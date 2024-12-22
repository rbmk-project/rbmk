
# rbmk pipe write - Write to Named Pipe

## Usage

```
rbmk pipe write PIPE
```

## Description

Write to a named pipe in the current directory. Blocks until a reader
connects or until the connection timeout expires (1s).

Data is read from stdin and written to the pipe.

We implement the named pipe using Unix domain sockets as documented in
the `rbmk pipe --help` output message.

## Examples

Write string to pipe:

```
$ echo "data" | rbmk pipe write mypipe
```

Write measurement results:

```
$ rbmk dig +short=ip example.com | rbmk pipe write addresses
```

## Exit Status

Returns `0` on success. Returns `1` on:

- Usage errors (missing pipe name)

- Connection timeout

- I/O errors

## Bugs

See `rbmk pipe --help` for details regarding issues that could arise
with the maximum platform-specific Unix domain socket path length.

## History

The `rbmk pipe write` command was introduced in RBMK v0.4.0.
