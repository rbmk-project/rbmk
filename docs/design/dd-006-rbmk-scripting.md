# RBMK Scripting Support Design Document

|              |                                                |
|--------------|------------------------------------------------|
| Author       | [@bassosimone](https://github.com/bassosimone) |
| Last-Updated | 2024-12-23                                     |

This document describes the scripting support design
for RBMK, which builds on top of the core functionality
described in [dd-005-rbmk.md](dd-005-rbmk.md).

## Overview

RBMK scripting support enables defining measurement
algorithms as shell scripts, with focus on remote deployment
and minimal dependencies.

## Core Principles

1. Scripts are the primary way to define measurement algorithms
2. Scripts should be self-contained and portable
3. Minimal dependencies (just rbmk binary + script)
4. POSIX shell compatibility across platforms

## Command Set

We extend RBMK with Unix-like commands to support scripting:

- `rbmk cat`: Concatenate files
- `rbmk head`: Print first lines of files
- `rbmk ipuniq`: Sort, deduplicate, and format IP addresses
- `rbmk markdown`: Renders Markdown to console
- `rbmk mkdir`: Create directories
- `rbmk mv`: Move (rename) files and directories
- `rbmk pipe`: Create named pipes for inter-process communication
- `rbmk rm`: Remove files and directories
- `rbmk sh`: POSIX shell interpreter with `RBMK_EXE` environment variable
- `rbmk tar`: Create and compress TAR archives
- `rbmk timestamp`: Generate filesystem-friendly timestamps

These commands work consistently across operating systems and handle
platform-specific details (e.g., path separators).

## Remote Measurement Workflow

1. Generate measurement script locally
2. Copy script and `rbmk` binary to target machine
3. Execute measurement (e.g., `./rbmk sh script.bash`)
4. Retrieve results

Example:

```bash
# Generate measurement script using your custom script generator
./your/custom/script/generator -i example.com >measure.sh

# Copy rbmk and script to remote machine
scp rbmk measure.sh user@remote:

# Execute script remotely
ssh -v user@remote
./rbmk sh measure.sh
```

The separation between generation and execution allows for more
operational agility in the measurement process. Scripts generation
can happen inside an uncensored environment, while execution has
minimal runtime requirements. When copying script files across
long-distance networks, the script files can be compressed, and
occupy much less bandwidth than the `rbmk` binary itself or
the measurement data. As a result, it will be possible to iterate
quickly on updating and re-running the measurement scripts.

## Script Execution

## Script Execution

The `rbmk` binary provides a POSIX shell through `rbmk sh` that guarantees
script portability across different operating systems. Key features:

1. **Restricted Command Set**

- Only `rbmk` commands are available within scripts as built-in commands
- External commands cannot be executed
- All core measurement and Unix-like operations are provided through `rbmk` subcommands

2. **Consistent Behavior**

- Scripts work identically on Unix-like systems and Windows
- Environment differences are abstracted away

3. **Development Workflow**

- Scripts can be developed and tested locally
- Deploy to any system without modification
- No dependency on external tools
- No surprises when running in different environments

Example script showing the contained environment:

```bash
#!/bin/sh
# This script will work identically everywhere using `./rbmk sh`
timestamp=$(rbmk timestamp)
rbmk mkdir -p "$timestamp"
rbmk dig +short=ip example.com > "$timestamp/addrs.txt"
rbmk tar -czf "results_$timestamp.tar.gz" "$timestamp"
rbmk rm -rf "$timestamp"
```

The restriction to `rbmk` commands ensures that scripts:

1. Are truly portable across systems
2. Have predictable behavior
3. Can be tested once and deployed anywhere
4. Don't break due to missing or different external tools

## Data Management

Scripts should use relative paths for data management. For example:

```bash
./results/      # measurement results
./logs/         # structured logs
```

This ensures consistency across different execution environments.

## Design Choices

1. Shell Interpreter
- Use `mvdan.cc/sh` for portable, consistent behavior
- Provide `RBMK_EXE` environment variable so that a script can
invoke `rbmk` subcommands

2. Script Generation
- Python (or other scripting languages) for generation logic
- Built-in shell for execution
- Clear separation of roles

3. Unix-like Commands
- Minimal but complete set
- Platform-independent behavior
- Focus on measurement needs

## Future Extensions

1. Script Templates
- Common measurement patterns
- Best practices examples
- Reusable components

2. Result Processing
- Standard analysis tools
- Report generation
- Data visualization

3. Unix Commands
- More commands as needed

## References

1. [RBMK Core Design](dd-005-rbmk.md)
2. [mvdan.cc/sh/v3 Documentation](https://pkg.go.dev/mvdan.cc/sh/v3)
