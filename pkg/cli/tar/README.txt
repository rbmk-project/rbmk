
usage: rbmk tar -czf ARCHIVE FILE|DIR...

Create a tar ARCHIVE containing the specified FILEs and DIRs. We
only support archiving regular files and directories.

We currently support the following command line flags:

    -c, --create
        Create a new archive.

    -f, --file NAME
        Archive file name.

    -h, --help
        Print this help message.

    -z, --gzip
        Compress the archive with gzip.

For example, the following command creates a compressed archive named
`results.tar.gz` containing the `measurements` directory contents:

    $ rbmk tar -czf results.tar.gz ./measurements

This command exits with `0` on success and `1` on failure.
