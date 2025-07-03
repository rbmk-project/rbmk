
# rbmk cat - File Concatenation

## Usage

```
rbmk cat [FILE...]
```

## Description

Concatenate files and print on the standard output. If no `FILE`
is specified, read from the standard input. If `FILE` is `-`,
read from the standard input.

## Examples

The following invocation concatenates the content of the
`file1.txt` and `file2.txt` files to the stdout:

```
$ rbmk cat file1.txt file2.txt
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

Support from reading from the standard input and for treating
`-` as the standard input was introduced in RBMK v0.12.0.

The `rbmk cat` command was introduced in RBMK v0.2.0.
