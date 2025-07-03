
# rbmk head - Print First Lines

## Usage

```
rbmk head [-n COUNT] [FILE...]
```

## Description

Print the first `COUNT` lines of each `FILE` to the standard
output. With no specified `FILE`, or when the `FILE` is `-`, read
the standard input.

## Flags

### `-h, --help`

Print this help message.

### `-n, --lines COUNT`

Print the first `COUNT` lines instead of the first 10. The
`COUNT` must be a non-negative integer.

## Examples

Print first 10 lines from stdin:

```
$ rbmk dig +short=ip example.com | rbmk head
```

Print first 2 lines from stdin:

```
$ rbmk dig +short=ip example.com | rbmk head -n 2
```

Print first 5 lines from multiple files:

```
$ rbmk head -n 5 file1.txt file2.txt
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## Bugs

When running a command such as:

```
$ rbmk head
```

we keep the `stdin` in line-oriented mode, which means that you
can edit the input before pressing enter. However, this also implies
that `^C` does not interrupt reading from the `stdin`, because
the terminal driver is blocked reading until the EOL. The symptom
of this would be:

```
$ rbmk head
^C
```

where the program does not exit. To exit, insert an explicit
EOL character (e.g., `^D` on Unix and `^Z` + `Return` on Windows).

## History

The `rbmk head` command was introduced in RBMK v0.11.0.
