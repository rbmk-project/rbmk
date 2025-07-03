//
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from: https://github.com/spf13/afero
//

package fsx

import (
	"io"
	"io/fs"
	"net"
	"os"
	"time"
)

// osFileWrapper wraps an [*os.File] to hide the methods that are
// not defined by the [File] interface. This allows us to restrict
// the operations actually possible with a file.
type osFileWrapper struct {
	filep io.ReadWriteCloser
}

var _ File = osFileWrapper{}

// Close implements [File].
func (fw osFileWrapper) Close() error {
	return fw.filep.Close()
}

// Read implements [File].
func (fw osFileWrapper) Read(buf []byte) (int, error) {
	return fw.filep.Read(buf)
}

// Write implements [File].
func (fw osFileWrapper) Write(data []byte) (int, error) {
	return fw.filep.Write(data)
}

// maybeWrapOSFile wraps an [*os.File] into an [osFileWrapper] if the file
// was successfully opened. Otherwise, it returns the error.
func maybeWrapOSFile(filep io.ReadWriteCloser, err error) (File, error) {
	switch {
	case err != nil:
		return nil, err
	default:
		return osFileWrapper{filep}, nil
	}
}

// OsFS implements [FS] using the standard [os] package.
//
// The zero value is ready to use.
type OsFS struct{}

var (
	osChmod       = os.Chmod
	osChown       = os.Chown
	osChtimes     = os.Chtimes
	osCreate      = os.Create
	netDialUnix   = net.DialUnix
	netListenUnix = net.ListenUnix
	osLstat       = os.Lstat
	osMkdir       = os.Mkdir
	osMkdirAll    = os.MkdirAll
	osOpen        = os.Open
	osOpenFile    = os.OpenFile
	osReadDir     = os.ReadDir
	osRemove      = os.Remove
	osRemoveAll   = os.RemoveAll
	osRename      = os.Rename
	osStat        = os.Stat
)

// Chmod implements [FS].
func (OsFS) Chmod(name string, mode fs.FileMode) error {
	return osChmod(name, mode)
}

// Chown implements [FS].
func (OsFS) Chown(name string, uid, gid int) error {
	return osChown(name, uid, gid)
}

// Chtimes implements [FS].
func (OsFS) Chtimes(name string, atime, mtime time.Time) error {
	return osChtimes(name, atime, mtime)
}

// Create implements [FS].
func (OsFS) Create(name string) (File, error) {
	return maybeWrapOSFile(osCreate(name))
}

// DialUnix implements [FS].
//
// See also the limitations documented in the top-level package docs.
func (OsFS) DialUnix(name string) (net.Conn, error) {
	return netDialUnix("unix", nil, &net.UnixAddr{Name: name, Net: "unix"})
}

// ListenUnix implements [FS].
//
// See also the limitations documented in the top-level package docs.
func (OsFS) ListenUnix(name string) (net.Listener, error) {
	return netListenUnix("unix", &net.UnixAddr{Name: name, Net: "unix"})
}

// Lstat implements [FS].
func (OsFS) Lstat(name string) (fs.FileInfo, error) {
	return osLstat(name)
}

// Mkdir implements [FS].
func (OsFS) Mkdir(name string, mode fs.FileMode) error {
	return osMkdir(name, mode)
}

// MkdirAll implements [FS].
func (OsFS) MkdirAll(name string, mode fs.FileMode) error {
	return osMkdirAll(name, mode)
}

// Open implements [FS].
func (OsFS) Open(name string) (File, error) {
	return maybeWrapOSFile(osOpen(name))
}

// OpenFile implements [FS].
func (OsFS) OpenFile(name string, flag int, mode fs.FileMode) (File, error) {
	return maybeWrapOSFile(osOpenFile(name, flag, mode))
}

// ReadDir implements [FS].
func (OsFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return osReadDir(name)
}

// Remove implements [FS].
func (OsFS) Remove(name string) error {
	return osRemove(name)
}

// RemoveAll implements [FS].
func (OsFS) RemoveAll(name string) error {
	return osRemoveAll(name)
}

// Rename implements [FS].
func (OsFS) Rename(oldname, newname string) error {
	return osRename(oldname, newname)
}

// Stat implements [FS].
func (OsFS) Stat(name string) (os.FileInfo, error) {
	return osStat(name)
}
