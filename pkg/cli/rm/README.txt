
usage: rbmk rm [-rf] file...

Remove files or directories.

We currently support the following command line flags:

    -f, --force
        Ignore nonexistent-file errors.

    -h, --help
        Print this help message.

    -r, --recursive
        Remove directories and their contents recursively.

For example, the following invocation removes `file1.txt` and `file2.txt`:

    $ rbmk rm file1.txt file2.txt

The following invocation removes a directory and its contents:

    $ rbmk rm -rf directory

This command exits with `0` on success and `1` on failure.
