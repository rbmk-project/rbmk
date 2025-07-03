# Common Design Document

|              |                                                |
|--------------|------------------------------------------------|
| Author       | [@bassosimone](https://github.com/bassosimone) |
| Last-Updated | 2024-11-13                                     |

This document describes the design of the [common](
https://github.com/rbmk-project/common) library collection,
which contains several general purpose libraries reused across
[rbmk](https://github.com/rbmk-project) repositories.

A library added to [common](https://github.com/rbmk-project/common)
should respect these criteria:

1. stable (no breaking changes);

2. well tested and well documented;

3. useful across 2+ repositories;

4. reasonably small.

Otherwise, it is just best to define libraries when they are needed
or to create new repositories for large libraries.
