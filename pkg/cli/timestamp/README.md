
# rbmk timestamp - UTC Timestamp Generation

## Usage

```
rbmk timestamp
```

## Description

Print a filesystem-friendly ISO8601 UTC timestamp.

The timestamp format is YYYYMMDDTHHmmssZ, for example:

    20241201T114117Z

## Features

This timestamp format:

- Is sortable (chronological order)

- Contains no spaces or special characters

- Is safe for use in filenames

- Uses UTC timezone (indicated by Z suffix)

- Follows the ISO 8601 compact format

## Examples

Create directory with timestamped name:

```
$ outdir="./Workspace/$(rbmk timestamp)"
$ rbmk mkdir -p "$outdir"
```

## Exit Status

This command exits with `0` on success and `1` on failure.