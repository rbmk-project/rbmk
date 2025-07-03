//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

package fsx

// NewRelativeFS is a deprecated alias for [NewContainedFS].
//
// Deprecated: use [NewContainedFS] instead.
var NewRelativeFS = NewContainedFS

// RelativeFS is a deprecated alias for [ContainedFS].
//
// Deprecated: use [ContainedFS] instead.
type RelativeFS = ContainedFS

// NewContainedFS creates a new [FS] rooted at the given path
// using the given child [FS] as the dependency.
//
// Any file name (after [filepath.Clean]) outside this base
// path will be treated as a non-existing file.
//
// Any absolute file name will be treated as a non-existing file.
//
// We return [fs.ErrNotExist] in these cases.
//
// Note: This implementation cannot prevent symlink traversal
// attacks. The caller must ensure the base directory does not
// contain symlinks if this is a security requirement.
//
// Deprecated: use [NewOverlayFS] with [NewRelativeContainedDirPathMapper] instead.
func NewContainedFS(dep FS, path string) *ContainedFS {
	return &ContainedFS{NewOverlayFS(dep, NewRelativeContainedDirPathMapper(path))}
}

// ContainedFS is the [FS] type returned by [NewContainedFS].
//
// The zero value IS NOT ready to use; construct using [NewContainedFS].
//
// Deprecated: use [NewOverlayFS] with [NewRelativeContainedDirPathMapper] instead.
type ContainedFS struct {
	*OverlayFS
}
