
usage: rbmk COMMAND [args...]

RBMK (Really Basic Measurement Kit) is a command-line utility
to facilitate network epxloration and measurements.

We support these commands:

    curl    Emulates a subset of the `curl(1)` command.
    dig     Emulates a subset of the `dig(1)` command.

You can combine these two commands as follows:

    $ rbmk dig +short example.com
    93.184.215.14

    $ rbmk curl --resolve example.com:443:93.184.215.14 https://example.com/
    <!doctype html>
    <html>
    [... rest of the response body ...]

Run `rbmk COMMAND --help` for more information about `COMMAND`.
