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
	"os"
	"path/filepath"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/x/closepool"
	"github.com/spf13/pflag"
)

//go:embed README.txt
var readme string

// NewCommand creates the `rbmk tar` Command.
func NewCommand() cliutils.Command {
	return command{}
}

type command struct{}

func (cmd command) Help(argv ...string) error {
	fmt.Fprintf(os.Stdout, "%s\n", readme)
	return nil
}

func (cmd command) Main(ctx context.Context, argv ...string) error {
	// 1. honour requests for printing the help
	if cliutils.HelpRequested(argv...) {
		return cmd.Help(argv...)
	}

	// 2. parse command line flags
	clip := pflag.NewFlagSet("rbmk tar", pflag.ContinueOnError)
	compress := clip.BoolP("gzip", "z", false, "compress archive with gzip")
	create := clip.BoolP("create", "c", false, "create a new archive")
	file := clip.StringP("file", "f", "", "archive file name")

	if err := clip.Parse(argv[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk tar --help` for usage.\n")
		return err
	}

	// 3. validate flags combination
	if !*create {
		err := errors.New("only archive creation is supported")
		fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
		return err
	}
	if *file == "" {
		err := errors.New("archive file name required")
		fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
		return err
	}

	// 4. ensure we have files to archive
	args := clip.Args()
	if len(args) < 1 {
		err := errors.New("no files to archive")
		fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Run `rbmk tar --help` for usage.\n")
		return err
	}

	// 5. Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// 6. create archive file
	filep, err := os.Create(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rbmk tar: cannot create archive: %s\n", err.Error())
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
		if err := appendToArchive(tw, path); err != nil {
			fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
			return err
		}
	}

	// 9. make sure everything is written to disk
	if err := pool.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "rbmk tar: %s\n", err.Error())
		return err
	}

	return nil
}

// appendToArchive adds a file or directory to the archive.
func appendToArchive(tw *tar.Writer, path string) error {
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

		// Create a tar header using the file info
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Ensure the name uses slashes
		header.Name = filepath.ToSlash(path)
		if info.IsDir() {
			header.Name += "/"
		}

		// Write the header to the archive
		if err := tw.WriteHeader(header); err != nil {
			return nil
		}

		// Attempt to copy the content of the file
		return copyFile(tw, info, path)
	})
}

// copyFile copies the content of a file to a tar writer.
func copyFile(tw *tar.Writer, info fs.FileInfo, filename string) error {
	if !info.Mode().IsRegular() {
		return nil
	}
	filep, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer filep.Close()
	_, err = io.Copy(tw, filep)
	return err
}
