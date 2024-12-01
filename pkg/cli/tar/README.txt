
usage: rbmk tar -czf archive.tar.gz files...

Create a tar archive containing the specified files and directories.

We currently support the following command line flags:

    -c, --create
        Create a new archive.

    -f, --file NAME
        Archive file name.

    -h, --help
        Print this help message.

    -z, --gzip
        Compress the archive with gzip.

For example:

    $ rbmk tar -czf results.tar.gz var/rbmk/measurements/
    Creates a compressed tar archive of the measurements directory.

This command exits with `0` on success and `1` on failure.
