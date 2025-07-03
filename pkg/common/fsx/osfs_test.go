// SPDX-License-Identifier: Apache-2.0

package fsx

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/mocks"
)

func TestOsFileWrapper(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		mockFile := &mocks.File{
			MockClose: func() error {
				return errors.New("close error")
			},
		}
		fileWrapper := osFileWrapper{filep: mockFile}

		err := fileWrapper.Close()
		if err == nil || err.Error() != "close error" {
			t.Errorf("expected close error, got %v", err)
		}
	})

	t.Run("Read", func(t *testing.T) {
		mockFile := &mocks.File{
			MockRead: func(b []byte) (int, error) {
				return 0, errors.New("read error")
			},
		}
		fileWrapper := osFileWrapper{filep: mockFile}

		buf := make([]byte, 5)
		n, err := fileWrapper.Read(buf)
		if err == nil || err.Error() != "read error" {
			t.Errorf("expected read error, got %v", err)
		}
		if n != 0 {
			t.Errorf("expected to read 0 bytes, read %d", n)
		}
	})

	t.Run("Write", func(t *testing.T) {
		mockFile := &mocks.File{
			MockWrite: func(b []byte) (int, error) {
				return 0, errors.New("write error")
			},
		}
		fileWrapper := osFileWrapper{filep: mockFile}

		data := []byte("hello, world")
		n, err := fileWrapper.Write(data)
		if err == nil || err.Error() != "write error" {
			t.Errorf("expected write error, got %v", err)
		}
		if n != 0 {
			t.Errorf("expected to write 0 bytes, wrote %d", n)
		}
	})
}

func TestMaybeWrapOSFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockFile := &mocks.File{}
		file, err := maybeWrapOSFile(mockFile, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if file == nil {
			t.Fatal("expected a file, got nil")
		}
	})

	t.Run("Failure", func(t *testing.T) {
		mockError := errors.New("open error")
		file, err := maybeWrapOSFile(nil, mockError)
		if err == nil || err.Error() != "open error" {
			t.Errorf("expected open error, got %v", err)
		}
		if file != nil {
			t.Errorf("expected nil file, got %v", file)
		}
	})
}

func TestOsFS(t *testing.T) {
	filesystem := OsFS{}

	t.Run("Chmod", func(t *testing.T) {
		original := osChmod
		defer func() { osChmod = original }()
		osChmod = func(name string, mode os.FileMode) error {
			return errors.New("chmod error")
		}

		err := filesystem.Chmod("dummy", 0644)
		if err == nil || err.Error() != "chmod error" {
			t.Errorf("expected chmod error, got %v", err)
		}
	})

	t.Run("Chown", func(t *testing.T) {
		original := osChown
		defer func() { osChown = original }()
		osChown = func(name string, uid, gid int) error {
			return errors.New("chown error")
		}

		err := filesystem.Chown("dummy", 0, 0)
		if err == nil || err.Error() != "chown error" {
			t.Errorf("expected chown error, got %v", err)
		}
	})

	t.Run("Chtimes", func(t *testing.T) {
		original := osChtimes
		defer func() { osChtimes = original }()
		osChtimes = func(name string, atime, mtime time.Time) error {
			return errors.New("chtimes error")
		}

		err := filesystem.Chtimes("dummy", time.Now(), time.Now())
		if err == nil || err.Error() != "chtimes error" {
			t.Errorf("expected chtimes error, got %v", err)
		}
	})

	t.Run("Create", func(t *testing.T) {
		original := osCreate
		defer func() { osCreate = original }()
		osCreate = func(name string) (*os.File, error) {
			return nil, errors.New("create error")
		}

		_, err := filesystem.Create("dummy")
		if err == nil || err.Error() != "create error" {
			t.Errorf("expected create error, got %v", err)
		}
	})

	t.Run("DialUnix", func(t *testing.T) {
		original := netDialUnix
		defer func() { netDialUnix = original }()
		netDialUnix = func(network string, laddr, raddr *net.UnixAddr) (*net.UnixConn, error) {
			return nil, errors.New("dial unix error")
		}

		_, err := filesystem.DialUnix("dummy")
		if err == nil || err.Error() != "dial unix error" {
			t.Errorf("expected dial unix error, got %v", err)
		}
	})

	t.Run("ListenUnix", func(t *testing.T) {
		original := netListenUnix
		defer func() { netListenUnix = original }()
		netListenUnix = func(network string, laddr *net.UnixAddr) (*net.UnixListener, error) {
			return nil, errors.New("listen unix error")
		}

		_, err := filesystem.ListenUnix("dummy")
		if err == nil || err.Error() != "listen unix error" {
			t.Errorf("expected listen unix error, got %v", err)
		}
	})

	t.Run("Lstat", func(t *testing.T) {
		original := osLstat
		defer func() { osLstat = original }()
		osLstat = func(name string) (os.FileInfo, error) {
			return nil, errors.New("lstat error")
		}

		_, err := filesystem.Lstat("dummy")
		if err == nil || err.Error() != "lstat error" {
			t.Errorf("expected lstat error, got %v", err)
		}
	})

	t.Run("Mkdir", func(t *testing.T) {
		original := osMkdir
		defer func() { osMkdir = original }()
		osMkdir = func(name string, perm os.FileMode) error {
			return errors.New("mkdir error")
		}

		err := filesystem.Mkdir("dummy", 0755)
		if err == nil || err.Error() != "mkdir error" {
			t.Errorf("expected mkdir error, got %v", err)
		}
	})

	t.Run("MkdirAll", func(t *testing.T) {
		original := osMkdirAll
		defer func() { osMkdirAll = original }()
		osMkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("mkdirall error")
		}

		err := filesystem.MkdirAll("dummy", 0755)
		if err == nil || err.Error() != "mkdirall error" {
			t.Errorf("expected mkdirall error, got %v", err)
		}
	})

	t.Run("Open", func(t *testing.T) {
		original := osOpen
		defer func() { osOpen = original }()
		osOpen = func(name string) (*os.File, error) {
			return nil, errors.New("open error")
		}

		_, err := filesystem.Open("dummy")
		if err == nil || err.Error() != "open error" {
			t.Errorf("expected open error, got %v", err)
		}
	})

	t.Run("OpenFile", func(t *testing.T) {
		original := osOpenFile
		defer func() { osOpenFile = original }()
		osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, errors.New("openfile error")
		}

		_, err := filesystem.OpenFile("dummy", os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil || err.Error() != "openfile error" {
			t.Errorf("expected openfile error, got %v", err)
		}
	})

	t.Run("ReadDir", func(t *testing.T) {
		original := osReadDir
		defer func() { osReadDir = original }()
		osReadDir = func(name string) ([]os.DirEntry, error) {
			return nil, errors.New("readdir error")
		}

		_, err := filesystem.ReadDir("dummy")
		if err == nil || err.Error() != "readdir error" {
			t.Errorf("expected readdir error, got %v", err)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		original := osRemove
		defer func() { osRemove = original }()
		osRemove = func(name string) error {
			return errors.New("remove error")
		}

		err := filesystem.Remove("dummy")
		if err == nil || err.Error() != "remove error" {
			t.Errorf("expected remove error, got %v", err)
		}
	})

	t.Run("RemoveAll", func(t *testing.T) {
		original := osRemoveAll
		defer func() { osRemoveAll = original }()
		osRemoveAll = func(path string) error {
			return errors.New("removeall error")
		}

		err := filesystem.RemoveAll("dummy")
		if err == nil || err.Error() != "removeall error" {
			t.Errorf("expected removeall error, got %v", err)
		}
	})

	t.Run("Rename", func(t *testing.T) {
		original := osRename
		defer func() { osRename = original }()
		osRename = func(oldpath, newpath string) error {
			return errors.New("rename error")
		}

		err := filesystem.Rename("old", "new")
		if err == nil || err.Error() != "rename error" {
			t.Errorf("expected rename error, got %v", err)
		}
	})

	t.Run("Stat", func(t *testing.T) {
		original := osStat
		defer func() { osStat = original }()
		osStat = func(name string) (os.FileInfo, error) {
			return nil, errors.New("stat error")
		}

		_, err := filesystem.Stat("dummy")
		if err == nil || err.Error() != "stat error" {
			t.Errorf("expected stat error, got %v", err)
		}
	})
}
