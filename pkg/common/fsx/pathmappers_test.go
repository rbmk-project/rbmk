// SPDX-License-Identifier: Apache-2.0

package fsx

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestPathMappers(t *testing.T) {
	curdir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	errMocked := errors.New("mocked error")

	type testCase struct {
		name      string
		construct func(baseDir string) (RealPathMapper, error)
		baseDir   string
		path      string
		mockAbs   func(string) (string, error)
		want      string
		wantError error
	}

	tests := []struct {
		group string
		cases []testCase
	}{
		{
			group: "PrefixDirPathMapper",
			cases: []testCase{
				{
					name: "absolute mapper with relative path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsolutePrefixDirPathMapper(baseDir)
					},
					baseDir: "testdata",
					path:    "file.txt",
					want:    filepath.Join(curdir, "testdata", "file.txt"),
				},

				{
					name: "absolute mapper with error",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsolutePrefixDirPathMapper(baseDir)
					},
					baseDir: "testdata",
					mockAbs: func(path string) (string, error) {
						return "", errMocked
					},
					wantError: errMocked,
				},

				{
					name: "relative mapper with relative path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewRelativePrefixDirPathMapper(baseDir), nil
					},
					baseDir: "testdata",
					path:    "file.txt",
					want:    filepath.Join("testdata", "file.txt"),
				},
			},
		},

		{
			group: "ContainedDirPathMapper",
			cases: []testCase{
				{
					name: "absolute mapper with relative path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsoluteContainedDirPathMapper(baseDir)
					},
					baseDir: "testdata",
					path:    "file.txt",
					want:    filepath.Join(curdir, "testdata", "file.txt"),
				},

				{
					name: "absolute mapper with absolute path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsoluteContainedDirPathMapper(baseDir)
					},
					baseDir:   "testdata",
					path:      "/file.txt",
					wantError: fs.ErrNotExist,
				},

				{
					name: "absolute mapper with outside path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsoluteContainedDirPathMapper(baseDir)
					},
					baseDir:   "testdata",
					path:      "../file.txt",
					wantError: fs.ErrNotExist,
				},

				{
					name: "absolute mapper with error",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewAbsoluteContainedDirPathMapper(baseDir)
					},
					baseDir: "testdata",
					mockAbs: func(path string) (string, error) {
						return "", errMocked
					},
					wantError: errMocked,
				},

				{
					name: "relative mapper with relative path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewRelativeContainedDirPathMapper(baseDir), nil
					},
					baseDir: "testdata",
					path:    "file.txt",
					want:    filepath.Join("testdata", "file.txt"),
				},

				{
					name: "relative mapper with absolute path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewRelativeContainedDirPathMapper(baseDir), nil
					},
					baseDir:   "testdata",
					path:      "/file.txt",
					wantError: fs.ErrNotExist,
				},

				{
					name: "relative mapper with outside path",
					construct: func(baseDir string) (RealPathMapper, error) {
						return NewRelativeContainedDirPathMapper(baseDir), nil
					},
					baseDir:   "testdata",
					path:      "../file.txt",
					wantError: fs.ErrNotExist,
				},
			},
		},
	}

	for _, group := range tests {
		t.Run(group.group, func(t *testing.T) {
			for _, tt := range group.cases {
				t.Run(tt.name, func(t *testing.T) {
					if tt.mockAbs != nil {
						saved := filepathAbs
						filepathAbs = tt.mockAbs
						defer func() { filepathAbs = saved }()
					}

					pmap, err := tt.construct(tt.baseDir)
					if err != nil {
						if !errors.Is(err, tt.wantError) {
							t.Fatalf("unexpected construction error: got %v, want %v", err, tt.wantError)
						}
						return
					}

					got, err := pmap.RealPath(tt.path)
					if !errors.Is(err, tt.wantError) {
						t.Fatalf("unexpected error: got %v, want %v", err, tt.wantError)
					}
					if err == nil && got != tt.want {
						t.Fatalf("got %q, want %q", got, tt.want)
					}
				})
			}
		})
	}
}

func TestRelativeToCwdPrefixDirPathMapper(t *testing.T) {
	type testCase struct {
		// name is the name of the test case
		name string

		// mockCwd is the mock to use for [os.Cwd]
		mockCwd func() (string, error)

		// mockRel is the mock to use for [filepath.Rel]
		mockRel func(cwd, path string) (string, error)

		// inputPath is the path passed to [NewRelativeToCwdPrefixDirPathMapper]
		inputPath string

		// want is the resulting directory that we want
		want string

		// wantError is the error that we expect
		wantError error
	}

	tests := []testCase{
		{
			name: "simple relative path",
			mockCwd: func() (string, error) {
				return "/base", nil
			},
			mockRel: func(cwd, path string) (string, error) {
				if cwd != "/base" || path != "/base/project" {
					t.Fatalf("unexpected args: cwd=%q path=%q", cwd, path)
				}
				return "project", nil
			},
			inputPath: "/base/project",
			want:      "project",
		},

		{
			name: "nested path",
			mockCwd: func() (string, error) {
				return "/base", nil
			},
			mockRel: func(cwd, path string) (string, error) {
				if cwd != "/base" || path != "/base/deep/project" {
					t.Fatalf("unexpected args: cwd=%q path=%q", cwd, path)
				}
				return "deep/project", nil
			},
			inputPath: "/base/deep/project",
			want:      "deep/project",
		},

		{
			name: "getwd fails",
			mockCwd: func() (string, error) {
				return "", errors.New("getwd error")
			},
			mockRel: func(cwd, path string) (string, error) {
				t.Fatal("rel should not be called")
				return "", nil
			},
			inputPath: "/any/path",
			wantError: errors.New("getwd error"),
		},

		{
			name: "rel fails",
			mockCwd: func() (string, error) {
				return "/base", nil
			},
			mockRel: func(cwd, path string) (string, error) {
				return "", errors.New("rel error")
			},
			inputPath: "/any/path",
			wantError: errors.New("rel error"),
		},

		{
			name: "path outside base",
			mockCwd: func() (string, error) {
				return "/base", nil
			},
			mockRel: func(cwd, path string) (string, error) {
				if cwd != "/base" || path != "/other/path" {
					t.Fatalf("unexpected args: cwd=%q path=%q", cwd, path)
				}
				return "../other/path", nil
			},
			inputPath: "/other/path",
			want:      "../other/path", // Note: this is allowed by PrefixDirPathMapper
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original functions
			savedGetwd := osGetwd
			savedRel := filepathRel
			defer func() {
				osGetwd = savedGetwd
				filepathRel = savedRel
			}()

			// Install mocks
			osGetwd = tt.mockCwd
			filepathRel = tt.mockRel

			// Run test
			mapper, err := NewRelativeToCwdPrefixDirPathMapper(tt.inputPath)
			if err != nil {
				if tt.wantError == nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if err.Error() != tt.wantError.Error() {
					t.Fatalf("got error %v, want %v", err, tt.wantError)
				}
				return
			}
			if tt.wantError != nil {
				t.Fatalf("expected error %v, got nil", tt.wantError)
			}

			// Test the mapper
			got, err := mapper.RealPath("file.txt")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := filepath.Join(tt.want, "file.txt")
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
