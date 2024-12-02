
usage: rbmk mv [-f] SOURCE... DESTINATION

Move (rename) SOURCE to DESTINATION. When moving multiple SOURCE files,
the DESTINATION must be an existing directory.

We currently support the following command line flags:

    -h, --help
        Print this help message.

For example, to move a file:
    $ rbmk mv source.txt destination.txt

To move multiple files into a directory:
    $ rbmk mv file1.txt file2.txt target_directory/

This command exits with `0` on success and `1` on failure.
