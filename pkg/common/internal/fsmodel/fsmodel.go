// SPDX-License-Identifier: GPL-3.0-or-later

// Package fsmodel provides an abstract file system model.
package fsmodel

import (
	"io"
	"io/fs"
	"net"
	"os"
	"time"
)

// Forward file system constants.
const (
	O_CREATE = os.O_CREATE
	O_RDONLY = os.O_RDONLY
	O_RDWR   = os.O_RDWR
	O_TRUNC  = os.O_TRUNC
	O_WRONLY = os.O_WRONLY
	O_APPEND = os.O_APPEND
)

// File represents a file in the filesystem.
//
// We use a simplified view of the full interface implemented by
// [os.File] to allow for easier mocking and testing.
type File io.ReadWriteCloser

// Ensure [*os.File] implements [File].
var _ File = &os.File{}

// FS is the filesystem interface.
//
// Any simulated or real filesystem should implement this interface.
type FS interface {
	// Chmod changes the mode of the named file to mode.
	Chmod(name string, mode fs.FileMode) error

	// Chown changes the uid and gid of the named file.
	Chown(name string, uid, gid int) error

	// Chtimes changes the access and modification times of the named file.
	Chtimes(name string, atime time.Time, mtime time.Time) error

	// Create creates a file in the filesystem, returning the file or an error.
	Create(name string) (File, error)

	// DialUnix connects to a Unix-domain socket using the given file name.
	DialUnix(name string) (net.Conn, error)

	// ListenUnix creates a listening Unix-domain socket using the given file name.
	ListenUnix(name string) (net.Listener, error)

	// Lstat is like Stat but does not follow symbolic links.
	Lstat(name string) (fs.FileInfo, error)

	// Mkdir creates a directory in the filesystem, possibly returning an error.
	Mkdir(name string, perm fs.FileMode) error

	// MkdirAll creates a directory path and all parents that does not exist yet.
	MkdirAll(path string, perm fs.FileMode) error

	// Open opens a file, returning it or an error, if any.
	Open(name string) (File, error)

	// OpenFile opens a file using the given flags and the given mode.
	OpenFile(name string, flag int, perm fs.FileMode) (File, error)

	// ReadDir reads and returns the content of a given directory.
	ReadDir(dirname string) ([]fs.DirEntry, error)

	// Remove removes a file identified by name, returning an error, if any.
	Remove(name string) error

	// RemoveAll removes a directory path and any children it contains. It
	// does not fail if the path does not exist (returns nil).
	RemoveAll(path string) error

	// Rename renames a file.
	Rename(oldname, newname string) error

	// Stat returns a FileInfo describing the named file, or an error.
	Stat(name string) (fs.FileInfo, error)
}
