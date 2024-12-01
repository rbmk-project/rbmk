
usage: rbmk rm [-rf] file...

Remove files or directories.

We currently support the following command line flags:

    -f, --force
        Ignore nonexistent files and never prompt.

    -h, --help
        Print this help message.

    -r, --recursive
        Remove directories and their contents recursively.

For example:

    $ rbmk rm file1.txt file2.txt
    Remove individual files.

    $ rbmk rm -rf directory/
    Recursively remove a directory and its contents.

This command exits with `0` on success and `1` on failure.
