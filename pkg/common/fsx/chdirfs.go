//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

package fsx

// NewChdirFS creates a new [FS] where each file name is
// prefixed with the given directory path.
//
// Deprecated: use [NewOverlayFS] with [NewRelativePrefixDirPathMapper] instead.
func NewChdirFS(dep FS, path string) *ChdirFS {
	return &ChdirFS{NewOverlayFS(dep, NewRelativePrefixDirPathMapper(path))}
}

// ChdirFS is the [FS] type returned by [NewChdirFS].
//
// The zero value IS NOT ready to use; construct using [NewChdirFS].
//
// Deprecated: use [NewOverlayFS] with [NewRelativePrefixDirPathMapper] instead.
type ChdirFS struct {
	*OverlayFS
}
