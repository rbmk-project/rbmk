
usage: rbmk ipuniq file...

Read IP addresses from files, sort them, and remove duplicates.

The command:
- Reads IPv4 and IPv6 addresses from input files
- Normalizes different textual representations
- Removes duplicates
- Sorts addresses (IPv4 addresses before IPv6)

For example:

    $ cat dig_A.txt dig_AAAA.txt | rbmk ipuniq
    192.0.2.1
    2001:db8::1

The output can be safely used in shell scripts:

    $ for ip in $(rbmk ipuniq ips.txt); do
        rbmk curl --resolve "example.com:443:${ip}" https://example.com/
    done

This command exits with `0` on success and `1` on failure.
