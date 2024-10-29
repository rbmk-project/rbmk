
usage: rbmk dig [-flags] [@SERVER] NAME [TYPE] [+options]

The `rbmk dig` command emulate a subset of the `dig(1)` command. By default, we
print output on the standard output emulating what `dig(1)` would print.

Command line flags start with the `-` character, while query-specific
options start with the `+` character, just like in `dig(1)`.

The optional `@SERVER` argument indicates the name server to use for the
query. If omitted, we use `8.8.8.8` as the resolver.

The mandatory `NAME` argument indicates the domain name to query.

The optional `TYPE` argument indicates the query type. If missing, we issue
a query for the `A` record type. We support these record types:

    - A: resolves the IPv4 addresses associated with a domain name;

    - AAAA: resolves the IPv6 addresses associated with a domain name;

    - CNAME: resolves the canonical name of a domain name;

    - HTTPS: resolves the ALPNs and possibly IP address associaed
      with a domain name;

We currently do not support any command line flags.

We currently do not support any query options.

For example, the following invocation resolves `www.example.com` IPv6 address
(i.e., `AAAA` records) using the `1.1.1.1` name server:

    $ rbmk dig @1.1.1.1 www.example.com AAAA

This command exits with `0` on success and `1` on failure.
