
Quick Introduction

RBMK is designed for composable network measurements. The two core commands
are `dig` for DNS queries and `curl` for HTTP(S) requests. You can combine
these to perform step-by-step measurements.

Basic Examples:

1. DNS resolution only:
    $ rbmk dig +short example.com
    93.184.215.14

2. HTTP fetch using a specific IP:
    $ rbmk curl --resolve example.com:443:93.184.215.14 https://example.com/

3. Combining commands to measure DNS and HTTP separately:
    $ IP=$(rbmk dig +short example.com|head -n1)
    $ rbmk curl --resolve example.com:443:$IP https://example.com/

4. Collecting measurement data:
    $ rbmk dig --logs dns.jsonl example.com
    $ rbmk curl --logs http.jsonl https://example.com/

This modular approach helps isolate different aspects of network measurements
and makes it easier to understand where issues occur.

Run `rbmk dig --help` or `rbmk curl --help` for more detailed information
about each command's options.

For a more detailed guide, try `rbmk tutorial`.
