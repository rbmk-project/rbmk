
usage: rbmk timestamp

Print a filesystem-friendly ISO8601 UTC timestamp.

The timestamp format is YYYYMMDDTHHmmssZ, for example:

    20241201T114117Z

This format:

    - Is sortable (chronological order)

    - Contains no spaces or special characters

    - Is safe for use in filenames

    - Uses UTC timezone (indicated by Z suffix)

    - Follows the ISO 8601 compact format

The following example shows how to use this command in scripts
to create directories with timestamped names:

    $ outdir="./Workspace/$(rbmk timestamp)"
    $ rbmk mkdir -p "$outdir"

This command exits with `0` on success and `1` on failure.
