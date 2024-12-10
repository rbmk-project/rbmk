//
// SPDX-License-Identifier: BSD-3-Clause
//
// Adapted from: https://github.com/golang/go/blob/go1.23.3/src/archive/tar/writer.go
//

// Package tar implements the `rbmk tar` command.
package tar

import (
	"archive/tar"
	"compress/gzip"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/common/closepool"
	"github.com/rbmk-project/rbmk/internal/markdown"
	"github.com/spf13/pflag"
)

//go:embed README.md
var readme string

// NewCommand creates the `rbmk tar` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(env cliutils.Environment, argv ...string) error {
	fmt.Fprintf(env.Stdout(), "%s\n", markdown.MaybeRender(readme))
	return nil
}

func (cmd command) Main(ctx context.Context, env cliutils.Environment, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(env, argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk tar", pflag.ContinueOnError)
	create := clip.BoolP("create", "c", false, "create a new archive")
	file := clip.StringP("file", "f", "", "archive file name")
	compress := clip.BoolP("gzip", "z", false, "compress archive with gzip")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk tar --help` for usage.\n")
		return err
	}

	// 3. validate flags combination
	if !*create {
		err := errors.New("only archive creation is supported")
		fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
		return err
	}
	if *file == "" {
		err := errors.New("archive file name required")
		fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
		return err
	}

	// 4. ensure we have files to archive
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("expected one or more file or dir paths to archive")
		fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
		fmt.Fprintf(env.Stderr(), "Run `rbmk tar --help` for usage.\n")
		return err
	}

	// 5. Create a pool containing closers so that we can
	// close the chained writers in reverse order and handle
	// possible I/O errors while closing them.
	pool := &closepool.Pool{}
	defer pool.Close()

	// 6. create archive file
	filep, err := env.FS().Create(*file)
	if err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk tar: cannot create archive: %s\n", err.Error())
		return err
	}
	pool.Add(filep)

	// 7. setup writers
	var w io.Writer = filep
	if *compress {
		gw := gzip.NewWriter(filep)
		pool.Add(gw)
		w = gw
	}
	tw := tar.NewWriter(w)
	pool.Add(tw)

	// 8. add each file/directory to the archive
	for _, path := range args {
		if err := appendToArchive(env, tw, path); err != nil {
			fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
			return err
		}
	}

	// 9. make sure everything is written to disk
	// correctly w/o any I/O errors
	if err := pool.Close(); err != nil {
		fmt.Fprintf(env.Stderr(), "rbmk tar: %s\n", err.Error())
		return err
	}
	return nil
}

// appendToArchive adds a file or directory to the archive.
func appendToArchive(env cliutils.Environment, tw *tar.Writer, path string) error {
	return filepath.WalkDir(path, func(path string, dentry fs.DirEntry, err error) error {
		// Return early in case there's a walk error
		if err != nil {
			return err
		}

		// Obtain information about the file
		info, err := dentry.Info()
		if err != nil {
			return err
		}

		// We only support directories and regular files.
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", path)
		}

		// Create a tar header using the file info but avoid
		// bothering with the name of the file since we're need
		// to convert to slashes and handle dirs manually.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Ensure the name uses slashes and append `/` at the
		// end of directory names as required by the tar specification.
		header.Name = filepath.ToSlash(path)
		if info.IsDir() {
			header.Name += "/"
		}

		// Write the header to the archive
		if err := tw.WriteHeader(header); err != nil {
			return nil
		}

		// For directories, it suffices to write the header
		if info.IsDir() {
			return nil
		}

		// Attempt to copy the content of the file
		return copyRegularFile(env, tw, path)
	})
}

// copyRegularFile copies the content of a regular file to a tar writer.
func copyRegularFile(env cliutils.Environment, tw *tar.Writer, filename string) error {
	filep, err := env.FS().Open(filename)
	if err != nil {
		return err
	}
	defer filep.Close()
	_, err = io.Copy(tw, filep)
	return err
}
