//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

package fsx

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// RealPathMapper maps a virtual file name to its real path name.
type RealPathMapper interface {
	RealPath(virtualPath string) (realPath string, err error)
}

// RealPathMapperFunc is a function type that implements [RealPathMapper].
type RealPathMapperFunc func(virtualPath string) (realPath string, err error)

// Ensure [RealPathMapperFunc] implements [RealPathMapper].
var _ RealPathMapper = RealPathMapperFunc(nil)

// RealPath implements [RealPathMapper].
func (fx RealPathMapperFunc) RealPath(virtualPath string) (realPath string, err error) {
	return fx(virtualPath)
}

// Mockable [filepath.Abs] function for testing.
var filepathAbs = filepath.Abs

// PrefixDirPathMapper is a [RealPathMapper] that prepends
// a base directory to the virtual path.
//
// The zero value is invalid. Use [NewRelativePrefixDirPathMapper],
// [NewRelativeToCwdPrefixDirPathMapper] or [NewAbsolutePrefixDirPathMapper]
// to construct a new instance.
type PrefixDirPathMapper struct {
	// baseDir is the base directory to prepend.
	baseDir string
}

// ChdirPathMapper is a deprecated alias for [PrefixDirPathMapper].
type ChdirPathMapper = PrefixDirPathMapper

// NewAbsolutePrefixDirPathMapper converts the given directory
// to an absolute path and, on success, returns a new
// [*PrefixDirPathMapper] instance. On failure, it returns and error.
//
// # Usage Considerations
//
// Use this constructor when you want your [*PrefixDirPathMapper] to
// be robust against concurrent invocations of [os.Chdir].
func NewAbsolutePrefixDirPathMapper(baseDir string) (*PrefixDirPathMapper, error) {
	absBaseDir, err := filepathAbs(baseDir)
	if err != nil {
		return nil, err
	}
	return &PrefixDirPathMapper{baseDir: absBaseDir}, nil
}

// NewAbsoluteChdirPathMapper is a deprecated alias for [NewAbsolutePrefixDirPathMapper].
var NewAbsoluteChdirPathMapper = NewAbsolutePrefixDirPathMapper

// NewRelativePrefixDirPathMapper returns a new [*PrefixDirPathMapper]
// instance without bothering to check if the given directory
// is relative or absolute.
//
// # Usage Considerations
//
// Use this constructor when you know your program is not going
// to invoke [os.Chdir] so you can avoid building potentially long
// paths that could break Unix domain sockets as documented in
// the top-level package documentation.
func NewRelativePrefixDirPathMapper(baseDir string) *PrefixDirPathMapper {
	return &PrefixDirPathMapper{baseDir: baseDir}
}

// osGetwd allows to mock [os.Getwd] in tests.
var osGetwd = os.Getwd

// filepathRel allows to mock [filepath.Rel] in tests.
var filepathRel = filepath.Rel

// NewRelativeToCwdPrefixDirPathMapper returns a [*PrefixDirPathMapper] in which
// the given base directory is made relative to the current working directory
// obtained using [os.Getwd] at the time of the call. On failure, it returns an error.
//
// # Usage Considerations
//
// Use this constructor when you know your program is not going
// to invoke [os.Chdir] so you can avoid building potentially long
// paths that could break Unix domain sockets as documented in
// the top-level package documentation.
//
// This constructor explicitly addresses the `rbmk sh` use case where
// [mvdan.cc/sh/v3/interp] provides us with the absolute path of the
// current working directory, subcommands run as goroutines, we cannot
// chdir because we're still in the same process, and we want to minimise
// the length of paths because of Unix domain sockets path limitations.
func NewRelativeToCwdPrefixDirPathMapper(path string) (*PrefixDirPathMapper, error) {
	cwd, err := osGetwd()
	if err != nil {
		return nil, err
	}
	relPath, err := filepathRel(cwd, path)
	if err != nil {
		return nil, err
	}
	return NewRelativePrefixDirPathMapper(relPath), nil
}

// NewRelativeChdirPathMapper is a deprecated alias for [NewRelativePrefixDirPathMapper].
var NewRelativeChdirPathMapper = NewRelativePrefixDirPathMapper

// Ensure [PrefixDirPathMapper] implements [RealPathMapper].
var _ RealPathMapper = &PrefixDirPathMapper{}

// RealPath implements [RealPathMapper].
func (b *PrefixDirPathMapper) RealPath(virtualPath string) (realPath string, err error) {
	return filepath.Join(b.baseDir, virtualPath), nil
}

// ContainedDirPathMapper is a [RealPathMapper] that prevents
// accessing file names outside of a given base directory.
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
// The zero value is invalid. Use [NewRelativeContainedDirPathMapper] or
// [NewAbsoluteContainedDirPathMapper] to construct a new instance.
type ContainedDirPathMapper struct {
	// baseDir is the base directory to contain.
	baseDir string
}

// NewAbsoluteContainedDirPathMapper converts the given directory
// to an absolute path and, on success, returns a new [*ContainedDirPathMapper]
// instance. On failure, it returns and error.
//
// # Usage Considerations
//
// Use this constructor when you want your [*ContainedDirPathMapper] to
// be robust against concurrent invocations of [os.Chdir].
func NewAbsoluteContainedDirPathMapper(baseDir string) (*ContainedDirPathMapper, error) {
	absBaseDir, err := filepathAbs(baseDir)
	if err != nil {
		return nil, err
	}
	return &ContainedDirPathMapper{baseDir: absBaseDir}, nil
}

// NewRelativeContainedDirPathMapper returns a new [*ContainedDirPathMapper]
// instance without bothering to check if the given directory
// is relative or absolute.
//
// # Usage Considerations
//
// Use this constructor when you know your program is not going
// to invoke [os.Chdir] so you can avoid building potentially long
// paths that could break Unix domain sockets as documented in
// the top-level package documentation.
func NewRelativeContainedDirPathMapper(baseDir string) *ContainedDirPathMapper {
	return &ContainedDirPathMapper{baseDir: baseDir}
}

// Ensure [ContainedDirPathMapper] implements [RealPathMapper].
var _ RealPathMapper = &ContainedDirPathMapper{}

// RealPath implements [RealPathMapper].
func (c *ContainedDirPathMapper) RealPath(virtualPath string) (realPath string, err error) {
	// 1. entirely reject absolute path names
	if filepath.IsAbs(virtualPath) {
		return "", fs.ErrNotExist
	}

	// 2. clean the path and make sure it is not outside the base path
	bpath := filepath.Clean(c.baseDir)
	fullpath := filepath.Clean(filepath.Join(bpath, virtualPath))
	if !strings.HasPrefix(fullpath, bpath) {
		return "", fs.ErrNotExist
	}
	return fullpath, nil
}
