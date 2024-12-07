
# rbmk rm - Remove Files

## Usage

```
rbmk rm [-rf] file...
```

## Description

Remove files or directories.

## Flags

### `-f, --force`

Ignore nonexistent-file errors.

### `-h, --help`

Print this help message.

### `-r, --recursive`

Remove directories and their contents recursively.

## Examples

Remove multiple files:

```
$ rbmk rm file1.txt file2.txt
```

Remove directory recursively:

```
$ rbmk rm -rf directory
```

## Exit Status

This command exits with `0` on success and `1` on failure.
