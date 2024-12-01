
usage: rbmk mkdir [-p] directory...

Create the DIRECTORY(ies), if they do not already exist.

We currently support the following command line flags:

    -h, --help
        Print this help message.

    -p, --parents
        Create parent directories as needed.

For example:

    $ rbmk mkdir dir1 dir2 dir3
    Creates three directories at the current level.

    $ rbmk mkdir -p a/long/path/of/dirs
    Creates all parent directories as needed.

This command exits with `0` on success and `1` on failure.
