// SPDX-License-Identifier: Apache-2.0

package fsx_test

import (
	"errors"
	"io/fs"
	"net"
	"testing"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/fsx"
	"github.com/rbmk-project/rbmk/pkg/common/mocks"
)

func TestChdirFS(t *testing.T) {
	expected := errors.New("mocked error")

	t.Run("Chmod", func(t *testing.T) {
		expectedPath := "/base/attrs.txt"
		mockFS := &mocks.FS{
			MockChmod: func(name string, mode fs.FileMode) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Chmod("attrs.txt", 0600)

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Chown", func(t *testing.T) {
		expectedPath := "/base/attrs.txt"
		mockFS := &mocks.FS{
			MockChown: func(name string, uid, gid int) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Chown("attrs.txt", 1000, 1000)

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Chtimes", func(t *testing.T) {
		expectedPath := "/base/attrs.txt"
		mockFS := &mocks.FS{
			MockChtimes: func(name string, atime time.Time, mtime time.Time) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Chtimes("attrs.txt", time.Now(), time.Now())

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Create", func(t *testing.T) {
		expectedPath := "/base/test.txt"
		mockFS := &mocks.FS{
			MockCreate: func(name string) (fsx.File, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.Create("test.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("DialUnix", func(t *testing.T) {
		expectedPath := "/base/test.sock"
		mockFS := &mocks.FS{
			MockDialUnix: func(name string) (net.Conn, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.DialUnix("test.sock")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("ListenUnix", func(t *testing.T) {
		expectedPath := "/base/test.sock"
		mockFS := &mocks.FS{
			MockListenUnix: func(name string) (net.Listener, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.ListenUnix("test.sock")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Lstat", func(t *testing.T) {
		expectedPath := "/base/attrs.txt"
		mockFS := &mocks.FS{
			MockLstat: func(name string) (fs.FileInfo, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.Lstat("attrs.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Mkdir", func(t *testing.T) {
		expectedPath := "/base/testdir"
		mockFS := &mocks.FS{
			MockMkdir: func(name string, perm fs.FileMode) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Mkdir("testdir", 0755)

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("MkdirAll", func(t *testing.T) {
		expectedPath := "/base/testdir"
		mockFS := &mocks.FS{
			MockMkdirAll: func(name string, perm fs.FileMode) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.MkdirAll("testdir", 0755)

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Open", func(t *testing.T) {
		expectedPath := "/base/test.txt"
		mockFS := &mocks.FS{
			MockOpen: func(name string) (fsx.File, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.Open("test.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("OpenFile", func(t *testing.T) {
		expectedPath := "/base/modes.txt"
		mockFS := &mocks.FS{
			MockOpenFile: func(name string, flag int, perm fs.FileMode) (fsx.File, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.OpenFile("modes.txt", fsx.O_CREATE|fsx.O_WRONLY, 0644)

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("ReadDir", func(t *testing.T) {
		expectedPath := "/base/testdir"
		mockFS := &mocks.FS{
			MockReadDir: func(dirname string) ([]fs.DirEntry, error) {
				if dirname != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, dirname)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.ReadDir("testdir")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		expectedPath := "/base/remove.txt"
		mockFS := &mocks.FS{
			MockRemove: func(name string) error {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Remove("remove.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("RemoveAll", func(t *testing.T) {
		expectedPath := "/base/removedir"
		mockFS := &mocks.FS{
			MockRemoveAll: func(path string) error {
				if path != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, path)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.RemoveAll("removedir")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Rename", func(t *testing.T) {
		expectedOldPath := "/base/old.txt"
		expectedNewPath := "/base/new.txt"
		mockFS := &mocks.FS{
			MockRename: func(oldname, newname string) error {
				if oldname != expectedOldPath || newname != expectedNewPath {
					t.Fatalf("expected paths %q and %q, got %q and %q", expectedOldPath, expectedNewPath, oldname, newname)
				}
				return expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		err := chdirFS.Rename("old.txt", "new.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})

	t.Run("Stat", func(t *testing.T) {
		expectedPath := "/base/attrs.txt"
		mockFS := &mocks.FS{
			MockStat: func(name string) (fs.FileInfo, error) {
				if name != expectedPath {
					t.Fatalf("expected path %q, got %q", expectedPath, name)
				}
				return nil, expected
			},
		}
		chdirFS := fsx.NewChdirFS(mockFS, "/base")

		_, err := chdirFS.Stat("attrs.txt")

		if !errors.Is(err, expected) {
			t.Fatal("unexpected error", err)
		}
	})
}
