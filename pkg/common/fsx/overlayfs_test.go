// SPDX-License-Identifier: Apache-2.0

package fsx_test

import (
	"errors"
	"io/fs"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/fsx"
	"github.com/rbmk-project/rbmk/pkg/common/mocks"
)

func TestOverlayFS(t *testing.T) {
	expected := errors.New("mocked error")

	type testCase struct {
		name     string
		path     string
		setup    func(*mocks.FS)
		testFunc func(*fsx.OverlayFS, string) error
		want     error
	}

	tests := map[string][]testCase{
		"Chmod": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockChmod = func(name string, mode fs.FileMode) error {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Chmod(path, 0600)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Chmod(path, 0600)
				},
				want: fs.ErrNotExist,
			},
		},

		"Chown": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockChown = func(name string, uid, gid int) error {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Chown(path, 1000, 1000)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Chown(path, 1000, 1000)
				},
				want: fs.ErrNotExist,
			},
		},

		"Chtimes": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockChtimes = func(name string, atime, mtime time.Time) error {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					now := time.Now()
					return fs.Chtimes(path, now, now)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					now := time.Now()
					return fs.Chtimes(path, now, now)
				},
				want: fs.ErrNotExist,
			},
		},

		"Create": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockCreate = func(name string) (fsx.File, error) {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Create(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Create(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"DialUnix": {
			{
				name: "WithinBase",
				path: "/base/socket.sock",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockDialUnix = func(name string) (net.Conn, error) {
						if name != "/base/socket.sock" {
							t.Fatalf("expected path %q, got %q", "/base/socket.sock", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.DialUnix(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.DialUnix(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"ListenUnix": {
			{
				name: "WithinBase",
				path: "/base/socket.sock",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockListenUnix = func(name string) (net.Listener, error) {
						if name != "/base/socket.sock" {
							t.Fatalf("expected path %q, got %q", "/base/socket.sock", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.ListenUnix(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.ListenUnix(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"Lstat": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockLstat = func(name string) (fs.FileInfo, error) {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Lstat(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Lstat(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"Mkdir": {
			{
				name: "WithinBase",
				path: "/base/dir",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockMkdir = func(name string, perm fs.FileMode) error {
						if name != "/base/dir" {
							t.Fatalf("expected path %q, got %q", "/base/dir", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Mkdir(path, 0755)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Mkdir(path, 0755)
				},
				want: fs.ErrNotExist,
			},
		},

		"MkdirAll": {
			{
				name: "WithinBase",
				path: "/base/dir/subdir",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockMkdirAll = func(name string, perm fs.FileMode) error {
						if name != "/base/dir/subdir" {
							t.Fatalf("expected path %q, got %q", "/base/dir/subdir", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.MkdirAll(path, 0755)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.MkdirAll(path, 0755)
				},
				want: fs.ErrNotExist,
			},
		},

		"Open": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockOpen = func(name string) (fsx.File, error) {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Open(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Open(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"OpenFile": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockOpenFile = func(name string, flag int, perm fs.FileMode) (fsx.File, error) {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.OpenFile(path, fsx.O_CREATE|fsx.O_WRONLY, 0644)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.OpenFile(path, fsx.O_CREATE|fsx.O_WRONLY, 0644)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"ReadDir": {
			{
				name: "WithinBase",
				path: "/base/dir",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockReadDir = func(name string) ([]fs.DirEntry, error) {
						if name != "/base/dir" {
							t.Fatalf("expected path %q, got %q", "/base/dir", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.ReadDir(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.ReadDir(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},

		"Remove": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockRemove = func(name string) error {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Remove(path)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Remove(path)
				},
				want: fs.ErrNotExist,
			},
		},

		"RemoveAll": {
			{
				name: "WithinBase",
				path: "/base/dir",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockRemoveAll = func(name string) error {
						if name != "/base/dir" {
							t.Fatalf("expected path %q, got %q", "/base/dir", name)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.RemoveAll(path)
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.RemoveAll(path)
				},
				want: fs.ErrNotExist,
			},
		},

		"Rename": {
			{
				name: "WithinBase",
				path: "/base/old.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockRename = func(oldname, newname string) error {
						if oldname != "/base/old.txt" || newname != "/base/new.txt" {
							t.Fatalf("expected paths %q and %q, got %q and %q", "/base/old.txt", "/base/new.txt", oldname, newname)
						}
						return expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Rename(path, "/base/new.txt")
				},
				want: expected,
			},

			{
				name:  "OutsideBaseFirst",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Rename(path, "/base/new.txt")
				},
				want: fs.ErrNotExist,
			},

			{
				name:  "OutsideBaseSecond",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					return fs.Rename("/base/old.txt", path)
				},
				want: fs.ErrNotExist,
			},
		},

		"Stat": {
			{
				name: "WithinBase",
				path: "/base/file.txt",
				setup: func(mockFS *mocks.FS) {
					mockFS.MockStat = func(name string) (fs.FileInfo, error) {
						if name != "/base/file.txt" {
							t.Fatalf("expected path %q, got %q", "/base/file.txt", name)
						}
						return nil, expected
					}
				},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Stat(path)
					return err
				},
				want: expected,
			},

			{
				name:  "OutsideBase",
				path:  "../outside",
				setup: func(mockFS *mocks.FS) {},
				testFunc: func(fs *fsx.OverlayFS, path string) error {
					_, err := fs.Stat(path)
					return err
				},
				want: fs.ErrNotExist,
			},
		},
	}

	for groupName, group := range tests {
		for _, tt := range group {
			t.Run(groupName+": "+tt.name, func(t *testing.T) {
				mockFS := &mocks.FS{}
				tt.setup(mockFS)

				// We create a real path mapper that rejects any path not
				// starting with `/base` to test both the code path in which
				// a path is accepted and the one in which it is rejected.
				mapper := fsx.RealPathMapperFunc(func(path string) (string, error) {
					if !strings.HasPrefix(path, "/base") {
						return "", fs.ErrNotExist
					}
					return path, nil
				})
				relativeFS := fsx.NewOverlayFS(mockFS, mapper)

				err := tt.testFunc(relativeFS, tt.path)

				if !errors.Is(err, tt.want) {
					t.Errorf("got error %v, want %v", err, tt.want)
				}
			})
		}
	}
}
