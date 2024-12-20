
# rbmk mkdir - Directory Creation

## Usage

```
rbmk mkdir [-p] DIRECTORY...
```

Create the `DIRECTORY`(ies), if they do not already exist. We use the
`0755` file mode to create new directories.

## Flags

### `-h, --help`

Print this help message.

### `-p, --parents`

Create parent directories as needed.

## Examples

Create multiple directories:

```
$ rbmk mkdir dir1 dir2 dir3
```

Create nested directories:

```
$ rbmk mkdir -p a/long/path/of/dirs
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

The `rbmk mkdir` command was introduced in RBMK v0.2.0.
