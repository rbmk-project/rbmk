//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

package fsx

import (
	"io/fs"
	"net"
	"time"
)

// OverlayFS overlays a path manipulation function implemented
// by [RealPathMapper] on top of another [FS].
//
// The zero value is invalid. Construct using [NewOverlayFS].
type OverlayFS struct {
	// rpm is the real path mapper.
	rpm RealPathMapper

	// fs is the underlying [FS].
	fs FS
}

// NewOverlayFS creates a new [FS] that uses [RealPathMapper] to
// map virtual paths to real paths on top of the given [FS].
func NewOverlayFS(fs FS, rpm RealPathMapper) *OverlayFS {
	return &OverlayFS{rpm: rpm, fs: fs}
}

// Ensure [OverlayFS] implements [FS].
var _ FS = &OverlayFS{}

// Chmod implements [FS].
func (rfs *OverlayFS) Chmod(name string, mode fs.FileMode) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "chmod", Path: name, Err: err}
	}
	return rfs.fs.Chmod(name, mode)
}

// Chown implements [FS].
func (rfs *OverlayFS) Chown(name string, uid, gid int) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "chown", Path: name, Err: err}
	}
	return rfs.fs.Chown(name, uid, gid)
}

// Chtimes implements [FS].
func (rfs *OverlayFS) Chtimes(name string, atime, mtime time.Time) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "chtimes", Path: name, Err: err}
	}
	return rfs.fs.Chtimes(name, atime, mtime)
}

// Create implements [FS].
func (rfs *OverlayFS) Create(name string) (File, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "create", Path: name, Err: err}
	}
	return rfs.fs.Create(name)
}

// DialUnix implements [FS].
//
// See also the limitations documented in the top-level package docs.
func (rfs *OverlayFS) DialUnix(name string) (net.Conn, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "dialunix", Path: name, Err: err}
	}
	return rfs.fs.DialUnix(name)
}

// ListenUnix implements [FS].
//
// See also the limitations documented in the top-level package docs.
func (rfs *OverlayFS) ListenUnix(name string) (net.Listener, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "listenunix", Path: name, Err: err}
	}
	return rfs.fs.ListenUnix(name)
}

// Lstat implements [FS].
func (rfs *OverlayFS) Lstat(name string) (fs.FileInfo, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: err}
	}
	return rfs.fs.Lstat(name)
}

// Mkdir implements [FS].
func (rfs *OverlayFS) Mkdir(name string, mode fs.FileMode) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return rfs.fs.Mkdir(name, mode)
}

// MkdirAll implements [FS].
func (rfs *OverlayFS) MkdirAll(name string, mode fs.FileMode) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "mkdir", Path: name, Err: err}
	}
	return rfs.fs.MkdirAll(name, mode)
}

// Open implements [FS].
func (rfs *OverlayFS) Open(name string) (File, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	return rfs.fs.Open(name)
}

// OpenFile implements [FS].
func (rfs *OverlayFS) OpenFile(name string, flag int, mode fs.FileMode) (File, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "openfile", Path: name, Err: err}
	}
	return rfs.fs.OpenFile(name, flag, mode)
}

// ReadDir implements [FS].
func (rfs *OverlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}
	return rfs.fs.ReadDir(name)
}

// Remove implements [FS].
func (rfs *OverlayFS) Remove(name string) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "remove", Path: name, Err: err}
	}
	return rfs.fs.Remove(name)
}

// RemoveAll implements [FS].
func (rfs *OverlayFS) RemoveAll(name string) error {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return &fs.PathError{Op: "removeall", Path: name, Err: err}
	}
	return rfs.fs.RemoveAll(name)
}

// Rename implements [FS].
func (rfs *OverlayFS) Rename(oldname, newname string) error {
	oldname, err := rfs.rpm.RealPath(oldname)
	if err != nil {
		return &fs.PathError{Op: "rename", Path: oldname, Err: err}
	}
	newname, err = rfs.rpm.RealPath(newname)
	if err != nil {
		return &fs.PathError{Op: "rename", Path: newname, Err: err}
	}
	return rfs.fs.Rename(oldname, newname)
}

// Stat implements [FS].
func (rfs *OverlayFS) Stat(name string) (fs.FileInfo, error) {
	name, err := rfs.rpm.RealPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}
	return rfs.fs.Stat(name)
}
