# Common Design Document

|              |                                                |
|--------------|------------------------------------------------|
| Author       | [@bassosimone](https://github.com/bassosimone) |
| Last-Updated | 2025-07-03                                     |

This document describes the design of the
[common](https://pkg.go.dev/github.com/rbmk-project/rbmk/pkg/common)
package, which contains simple, general purpose packages. Historically,
`common` was an independent repository, but we merged it into the main
RBMK repository on 2025-07-03.

A package added to
[common](https://pkg.go.dev/github.com/rbmk-project/rbmk/pkg/common)
should respect these criteria:

1. stable (no breaking changes);

2. well tested and well documented;

3. useful across 2+ repositories;

4. reasonably small;

5. should only depend on third party dependencies or on packages
that are already in [common](https://pkg.go.dev/github.com/rbmk-project/rbmk/pkg/common).

Otherwise, it is just best to define packages under other
package categories (e.g., `./pkg/dns`, `./pkg/x`).
