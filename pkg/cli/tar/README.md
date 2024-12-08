
# rbmk tar - Archive Creation

## Usage

```
rbmk tar -czf ARCHIVE FILE|DIR...
```

## Description

Create a tar ARCHIVE containing the specified FILEs and DIRs. We
only support archiving regular files and directories.

## Flags

### `-c, --create`

Create a new archive.

### `-f, --file NAME`

Set the archive file name.

### `-h, --help`

Print this help message.

### `-z, --gzip`

Compress the archive with gzip.

## Examples

Create a compressed archive named `results.tar.gz` containing the
`measurements` directory contents:

```
$ rbmk tar -czf results.tar.gz ./measurements
```

## Exit Status

This command exits with `0` on success and `1` on failure.
