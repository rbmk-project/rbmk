
usage: rbmk dig [-flags] [@SERVER] NAME [TYPE] [+options]

The `rbmk dig` command emulate a subset of the `dig(1)` command. By default, we
print output on the standard output emulating what `dig(1)` would print.

Command line flags start with the `-` character, while query-specific
options start with the `+` character, just like in `dig(1)`.

Flags MUST come first. The relative order of `@SERVER`, `NAME`, `TYPE`, and
`+options` is not significant, as long as they come after the flags.

The optional `@SERVER` argument indicates the name server to use for the
query. If omitted, we use `8.8.8.8` as the resolver. If `@SERVER` is specified
multiple times, we emit a warning and use the last one.

The mandatory `NAME` argument indicates the domain name to query. We do
not support specifying the `NAME` argument more than once.

The optional `TYPE` argument indicates the query type. If missing, we issue
a query for the `A` record type. We support these record types:

    - A: resolves the IPv4 addresses associated with a domain name;

    - AAAA: resolves the IPv6 addresses associated with a domain name;

    - CNAME: resolves the canonical name of a domain name;

    - HTTPS: resolves the ALPNs and possibly IP address associated
      with a domain name;

    - MX: resolves the mail exchange servers associated with a domain name;

    - NS: resolves the name servers associated with a domain name.

If you specify `TYPE` multiple times, we emit a warning and use the last one.

We currently support the following command line flags:

    -h, --help
        Print this help message.

    --logs FILE
        Writes structured logs to the given FILE. If FILE already exists, we
        append to it. If FILE does not exist, we create it. If FILE is a single
        dash (`-`), we write to the stdout. If you specify `--logs` multiple
        times, we write to the last FILE specified.

We currently support the following query options:

    +https
        Uses DNS-over-HTTPS. The @server argument is the hostname or IP
        address to use. The implied port is `443/tcp`. The implied URL
        path is `/dns-query`. That is, if you use:

            @8.8.8.8 +https

        We use `https://8.8.8.8/dns-query` to resolve the domain name.

    +logs
        Prints to the stdout structured logs showing network events
        occurred during the DNS resolution.

    +noall
        Suppress printing to the stdout.

    +qr
        Prints the query to the stdout before sending it.

    +short
        Print a short response rather than the full response.

    +tcp
        Uses DNS-over-TCP. The @server argument is the hostname or IP
        address to use. The implied port is `53/tcp`.

    +tls
        Uses DNS-over-TLS. The @server argument is the hostname or IP
        address to use. The implied port is `853/tcp`.

For example, the following invocation resolves `www.example.com` IPv6 address
(i.e., `AAAA` records) using the `1.1.1.1` name server:

    $ rbmk dig @1.1.1.1 www.example.com AAAA

To only print structured logs use `+noall +logs`:

    $ rbmk dig www.example.com MX +noall +logs

To append structured logs to a separate file, use the `--logs` flag:

    $ rbmk dig --logs LOGS.jsonl www.example.com MX

This command exits with `0` on success and `1` on failure.
