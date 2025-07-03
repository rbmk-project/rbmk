//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

/*
Package fsx allows abstracting the file system.

This package is derived from [afero].

In addition to [afero], we also implement support for dialing
and listening unix domain sockets, and for Lstat.

# Unix Domain Sockets Portability Note

Beware that Unix domain sockets have path length limitations ranging
around ~100 chars (e.g., 108 on Linux, 104 on macOS) for historical
reasons (see, e.g., https://unix.stackexchange.com/q/367008). When using
Unix domain sockets, therefore, it is possible to get `EINVAL` errors
in `bind` or `connect`, which occur when the combined path is too long. If
possible, consider using a relative base path (if you know you'll never
chdir elsewhere), or possibly use secure temporary directories.

[afero]: https://github.com/spf13/afero
*/
package fsx

import (
	"errors"
	"io/fs"
	"os"

	"github.com/rbmk-project/rbmk/pkg/common/internal/fsmodel"
)

// Forward file system constants.
const (
	O_CREATE = fsmodel.O_CREATE
	O_RDONLY = fsmodel.O_RDONLY
	O_RDWR   = fsmodel.O_RDWR
	O_TRUNC  = fsmodel.O_TRUNC
	O_WRONLY = fsmodel.O_WRONLY
	O_APPEND = fsmodel.O_APPEND
)

// IsNotExist combines the [os.IsNotExist] check with
// checking for the [fs.ErrNotExist] error.
func IsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err)
}

// File is an alias for [fsmodel.File].
type File = fsmodel.File

// Ensure [*os.File] implements [File].
var _ File = &os.File{}

// FS is an alias for [fsmodel.FS].
type FS = fsmodel.FS
