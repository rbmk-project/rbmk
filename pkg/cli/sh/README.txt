
usage: rbmk sh script

Run a shell script using a POSIX-compliant shell interpreter.

This shell implementation is consistent across platforms and supports:

- Variables and arithmetic

- Command substitution $(...)

- Pipes and redirections

- Loops and conditionals

- Environment variables

Environment:
    RBMK_EXE
        Automatically set to the absolute path of the rbmk executable.
        Use this in scripts to ensure rbmk commands work correctly
        regardless of the current working directory.

Example script:

    #!/usr/bin/env rbmk sh

    # Create measurement directory
    timestamp=$("${RBMK_EXE}" timestamp)
    outdir="var/rbmk/measurements/$timestamp"
    "${RBMK_EXE}" mkdir -p "$outdir"

    # Perform measurements
    "${RBMK_EXE}" dig +short=ip "dns.google" > "$outdir/dig.txt"

    # Archive results
    "${RBMK_EXE}" tar -czf "results_$timestamp.tar.gz" "$outdir"

This command exits with `0` on success and `1` on failure.
