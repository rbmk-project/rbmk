
# rbmk sh - Shell Script Execution

## Usage

```
rbmk sh SCRIPT [ARGUMENTS...]
```

## Description

Run `SCRIPT` using a POSIX-compliant shell interpreter providing
to the script the given `ARGUMENTS`, which will be available to
the script as `$1`, `$2`, etc.

This shell implementation (based on `mvdan.cc/sh/v3`) is consistent
across operating systems and supports:

- Variables and arithmetic.

- Command substitution `$(...)`.

- Pipes and redirections.

- Loops and conditionals

- Environment variables

## Available Commands

Apart from built-in commands (e.g., `cd`, `test`), the shell will
only allow running the `rbmk` command, which will behave as when you
normaly execute `rbmk`, except that `rbmk sh` won't be available.

We do this to restrict the set of commands that `rbmk sh` could run
and ensure scripts are portable. If you have more complex measurement
needs, we recommend using GNU bash instead.

## Environment

The `rbmk sh` command inherits the parent environment and includes the
following environment variables:

### `RBMK_EXE`

Automatically set to `rbmk` to allow scripts written before RBMK
v0.7.0 to continue running without modification. Since v0.7.0, `rbmk sh`
cannot execute external commands and is only allowed to run shell
built-in commands and the `rbmk` command.

## Example

The following example demonstrates how to use `rbmk sh` to run a script that:

1. creates a directory using a timestamp based name

2. uses `rbmk dig` to get the IP addresses of `dns.google`

3. archives the results into a tarball

4. removes the directory

First, let's see the content of the the `script.bash` file:

```sh
#!/bin/bash
set -x
outdir="$(rbmk timestamp --full)-$(rbmk random)"
rbmk mkdir -p "$outdir"
rbmk dig +short=ip A "dns.google" > "$outdir/dig1.txt"
rbmk dig +short=ip AAAA "dns.google" > "$outdir/dig2.txt"
rbmk tar -czf "results_$outdir.tar.gz" "$outdir"
rbmk rm -rf "$outdir"
```

To execute the script using `rbmk sh` run:

```
$ rbmk sh script.bash
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## Bugs

The `rbmk sh` command executes other `rbmk COMMAND` commands in the same
process without changing the current working directory. Because `rbmk pipe`
uses Unix domain sockets, and because such sockets are restricted in the
maximum path length, `rbmk sh` attempts to minimise path lengths by using
paths relative to the current working directory.

If `rbmk sh` cannot determine the current working directory when executing
a command, or it cannot map the subcommand own working directory to the
current working directory, it emits this warning message:

```
<date> rmbk sh: cannot create relative-to-cwd dir-path mapper: <error>
```

and then uses the absolute path instead.

Regardless, using long paths could lead to errors when using `rbmk pipe`
as documented in `rbmk pipe --help`. See `rbmk pipe --help` for advices
regarding mitigating this issue in shell scripts.

Note that, because `bash` changes the current working directory, it is
possible for scripts that work using `bash $script` to instead fail when
using `rbmk sh $script` precisely because of this issue.

## History

Since RBMK v0.10.0, it is possible to pass arguments to the script
executed by `rbmk sh` using the command line.

Before RBMK v0.7.0, `rbmk sh` used to set the `$RBMK_EXE` environment
variable to the `rbmk` path, to allow a script to execute `rbmk` commands.

Since v0.7.0. `rbmk` is an internal shell command, `rbmk sh` is not capable
of executing external commands, and `$RBMK_EXE` is set to `rbmk`, thus
supporting previously existing scripts without modification.

The `rbmk sh` command was introduced in RBMK v0.2.0.
