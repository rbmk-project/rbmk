# rbmk mv - Move Files

## Usage

```
rbmk mv [-f] SOURCE... DESTINATION
```

## Description

Move (rename) `SOURCE` to `DESTINATION`. When moving multiple `SOURCE` files,
the `DESTINATION` must be an existing directory.

## Flags

### `-h, --help`

Print this help message.

## Examples

Move a single file:

```
$ rbmk mv source.txt destination.txt
```

Move multiple files to a directory:

```
$ rbmk mv file1.txt file2.txt target_directory/
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

The `rbmk mv` command was introduced in RBMK v0.3.0.
