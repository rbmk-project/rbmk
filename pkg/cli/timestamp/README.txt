
usage: rbmk timestamp

Print a filesystem-friendly UTC timestamp.

The timestamp format is YYYYMMDDTHHmmssZ, for example:
20241130T223000Z

This format:
- Is sortable (chronological order)
- Contains no spaces or special characters
- Is safe for use in filenames
- Uses UTC timezone (indicated by Z suffix)
- Follows ISO 8601 compact format

For example:

    $ outdir="var/rbmk/measurements/$(rbmk timestamp)"
    $ rbmk mkdir -p "$outdir"

This command exits with `0` on success and `1` on failure.
