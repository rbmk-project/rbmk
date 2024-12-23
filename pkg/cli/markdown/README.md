
# rbmk markdown - Markdown Rendering

## Usage

```
rbmk markdown [flags]
```

## Description

Read markdown from standard input and write formatted output to standard
output. This command allows you to format markdown text consistently with
RBMK's own help texts.

Note: when RBMK is compiled with the `rbmk_disable_markdown` build tag,
this command prints its input unchanged to allow scripts to work even
when markdown support is disabled.

## Flags

### `-h, --help`

Print this help message.

## Examples

Basic usage:

```
$ echo "# Hello" | rbmk markdown
```

Format a help message in a script:

```bash
rbmk cat << 'EOF' | rbmk markdown
# My Script

This script does something useful.

## Usage

\`\`\`bash
./script.sh [flags]
\`\`\`

## Flags

### --option VALUE

Description of the option.
EOF
```

Format a README file:

```
$ rbmk cat README.md | rbmk markdown
```

## Exit Status

This command exits with `0` on success and `1` on failure.

## History

The `rbmk markdown` command was introduced in RBMK v0.12.0.
