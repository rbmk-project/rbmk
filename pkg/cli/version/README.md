
# rbmk version - Print Version Information

## Usage

```
rbmk version
```

## Description

Prints on the stdout version information. We add version information
when compiling `rbmk` using the `GNUMakefile`.

Possible values for the version information are:

- `dev` if we did not compile using the `GNUMakefile`.

- `vX.Y.Z` if using `GNUMakefile` to build a specific tag.

- `vX.Y.Z-<N>-g<SHA>` if using `GNUMakefile` to build
a commit not associated with a tag.
