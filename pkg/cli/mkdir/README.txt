
usage: rbmk mkdir [-p] DIRECTORY...

Create the DIRECTORY(ies), if they do not already exist. We use the
`0755` file mode to create new directories.

We currently support the following command line flags:

    -h, --help
        Print this help message.

    -p, --parents
        Create parent directories as needed.

For example, the following command creates three directories:

    $ rbmk mkdir dir1 dir2 dir3

The following command creates all parent directories as needed:

    $ rbmk mkdir -p a/long/path/of/dirs

This command exits with `0` on success and `1` on failure.
