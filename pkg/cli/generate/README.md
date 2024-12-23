
# rbmk generate - Script Generation

## Usage

```
rbmk generate TEMPLATE [flags]
```

## Description

Generate measurement scripts from templates. The generated scripts follow
RBMK best practices for error handling, logging, and progress tracking.

## Available Templates

* `stun_lookup` - STUN lookup measurement script

## Global Flags

### `-h, --help`

Print this help message.

## Examples

Generate a STUN lookup script:

```
$ rbmk generate stun_lookup > script.sh
```

Generate a STUN lookup script for specific endpoints:

```
$ rbmk generate stun_lookup \
    --input stun.l.google.com:19302 \
    --input stun.ekiga.net:3478 > script.sh
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

The `rbmk generate` command was introduced in RBMK v0.12.0.
