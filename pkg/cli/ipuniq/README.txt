
usage: rbmk ipuniq FILE...

Read IP addresses from files, sort them, and remove duplicates. We expect
input files to contain one IP address per line.

More specifically, `rbmk ipuniq`:

    - Reads IPv4 and IPv6 addresses from input files

    - Normalizes different textual representations

    - Removes duplicates

    - Randomly shuffles the resulting addresses

This command is useful to process the output of several
`rbmk dig` invocations that return IP addresses to create
a unique list of IP addresses to measure. For example:

    $ rbmk dig +short=ip example.com A > dig_A.txt

    $ rbmk dig + short=ip example.com AAAA > dig_AAAA.txt

    $ for ipAddr in $(rbmk ipuniq dig_A.txt dig_AAAA.txt); do \
        rbmk curl --resolve "example.com:443:${ipAddr}" https://example.com/ \
    done

This command exits with `0` on success and `1` on failure.
