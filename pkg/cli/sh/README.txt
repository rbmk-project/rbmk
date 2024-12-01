
usage: rbmk sh SCRIPT

Run SCRIPT using a POSIX-compliant shell interpreter.

This shell implementation (based on `mvdan.cc/sh/v3`) is consistent
across operating systems and supports:

    - Variables and arithmetic.

    - Command substitution `$(...)`.

    - Pipes and redirections.

    - Loops and conditionals

    - Environment variables

The `rbmk sh` command inherits the parent environment and includes the
following environment variables:

    RBMK_EXE
        Automatically set to the absolute path of the `rbmk` executable to
        help the script invoke `rbmk` commands.

The following example demonstrates how to use `rbmk sh` to run a script
that (1) creates a directory using a timestamp based name, (2) uses `rbmk
dig` to get the IP addresses of `dns.google`, (3) archives the results
into a tarball, (4) removes the directory:

    $ cat script.bash
    set -x
    timestamp=$("${RBMK_EXE}" timestamp)
    outdir="$timestamp"
    "${RBMK_EXE}" mkdir -p "$outdir"
    "${RBMK_EXE}" dig +short=ip A "dns.google" > "$outdir/dig1.txt"
    "${RBMK_EXE}" dig +short=ip AAAA "dns.google" > "$outdir/dig2.txt"
    "${RBMK_EXE}" tar -czf "results_$timestamp.tar.gz" "$outdir"
    "${RBMK_EXE}" rm -rf "$outdir"

    $ rbmk sh script.bash

This command exits with `0` on success and `1` on failure.
