package mocks

import (
	"io/fs"
	"net"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/internal/fsmodel"
)

// FsmodelFS is an alias for [fsmodel.FS].
type FsmodelFS = fsmodel.FS

// FsmodelFile is an alias for [fsmodel.File].
type FsmodelFile = fsmodel.File

// FS implements [FsmodelFS] for testing
type FS struct {
	// MockChmod implements Chmod
	MockChmod func(name string, mode fs.FileMode) error

	// MockChown implements Chown
	MockChown func(name string, uid, gid int) error

	// MockChtimes implements Chtimes
	MockChtimes func(name string, atime time.Time, mtime time.Time) error

	// MockCreate implements Create
	MockCreate func(name string) (FsmodelFile, error)

	// MockDialUnix implements DialUnix
	MockDialUnix func(name string) (net.Conn, error)

	// MockListenUnix implements ListenUnix
	MockListenUnix func(name string) (net.Listener, error)

	// MockLstat implements Lstat
	MockLstat func(name string) (fs.FileInfo, error)

	// MockMkdir implements Mkdir
	MockMkdir func(name string, perm fs.FileMode) error

	// MockMkdirAll implements MkdirAll
	MockMkdirAll func(path string, perm fs.FileMode) error

	// MockOpen implements Open
	MockOpen func(name string) (FsmodelFile, error)

	// MockOpenFile implements OpenFile
	MockOpenFile func(name string, flag int, perm fs.FileMode) (FsmodelFile, error)

	// MockReadDir implements ReadDir
	MockReadDir func(dirname string) ([]fs.DirEntry, error)

	// MockRemove implements Remove
	MockRemove func(name string) error

	// MockRemoveAll implements RemoveAll
	MockRemoveAll func(path string) error

	// MockRename implements Rename
	MockRename func(oldname, newname string) error

	// MockStat implements Stat
	MockStat func(name string) (fs.FileInfo, error)
}

// Ensure [FS] implements [FsmodelFS]
var _ FsmodelFS = &FS{}

// Chmod calls MockChmod
func (m *FS) Chmod(name string, mode fs.FileMode) error {
	return m.MockChmod(name, mode)
}

// Chown calls MockChown
func (m *FS) Chown(name string, uid, gid int) error {
	return m.MockChown(name, uid, gid)
}

// Chtimes calls MockChtimes
func (m *FS) Chtimes(name string, atime, mtime time.Time) error {
	return m.MockChtimes(name, atime, mtime)
}

// Create calls MockCreate
func (m *FS) Create(name string) (FsmodelFile, error) {
	return m.MockCreate(name)
}

// DialUnix calls MockDialUnix
func (m *FS) DialUnix(name string) (net.Conn, error) {
	return m.MockDialUnix(name)
}

// ListenUnix calls MockListenUnix
func (m *FS) ListenUnix(name string) (net.Listener, error) {
	return m.MockListenUnix(name)
}

// Lstat calls MockLstat
func (m *FS) Lstat(name string) (fs.FileInfo, error) {
	return m.MockLstat(name)
}

// Mkdir calls MockMkdir
func (m *FS) Mkdir(name string, perm fs.FileMode) error {
	return m.MockMkdir(name, perm)
}

// MkdirAll calls MockMkdirAll
func (m *FS) MkdirAll(path string, perm fs.FileMode) error {
	return m.MockMkdirAll(path, perm)
}

// Open calls MockOpen
func (m *FS) Open(name string) (FsmodelFile, error) {
	return m.MockOpen(name)
}

// OpenFile calls MockOpenFile
func (m *FS) OpenFile(name string, flag int, perm fs.FileMode) (FsmodelFile, error) {
	return m.MockOpenFile(name, flag, perm)
}

// ReadDir calls MockReadDir
func (m *FS) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return m.MockReadDir(dirname)
}

// Remove calls MockRemove
func (m *FS) Remove(name string) error {
	return m.MockRemove(name)
}

// RemoveAll calls MockRemoveAll
func (m *FS) RemoveAll(path string) error {
	return m.MockRemoveAll(path)
}

// Rename calls MockRename
func (m *FS) Rename(oldname, newname string) error {
	return m.MockRename(oldname, newname)
}

// Stat calls MockStat
func (m *FS) Stat(name string) (fs.FileInfo, error) {
	return m.MockStat(name)
}

// File implements [FsmodelFile] for testing
type File struct {
	// MockRead implements Read
	MockRead func(b []byte) (int, error)

	// MockWrite implements Write
	MockWrite func(b []byte) (int, error)

	// MockClose implements Close
	MockClose func() error
}

// Ensure [File] implements [FsmodelFile].
var _ FsmodelFile = &File{}

// Read calls MockRead
func (m *File) Read(b []byte) (int, error) {
	return m.MockRead(b)
}

// Write calls MockWrite
func (m *File) Write(b []byte) (int, error) {
	return m.MockWrite(b)
}

// Close calls MockClose
func (m *File) Close() error {
	return m.MockClose()
}
