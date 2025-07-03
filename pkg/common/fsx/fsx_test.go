// SPDX-License-Identifier: Apache-2.0

package fsx_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/rbmk-project/rbmk/pkg/common/fsx"
)

func TestIsNotExist(t *testing.T) {
	t.Run("fs.ErrNotExist", func(t *testing.T) {
		err := fs.ErrNotExist
		if !fsx.IsNotExist(err) {
			t.Fatal("expected true, got false")
		}
	})

	t.Run("os.ErrNotExist", func(t *testing.T) {
		err := os.ErrNotExist
		if !fsx.IsNotExist(err) {
			t.Fatal("expected true, got false")
		}
	})

	t.Run("other error", func(t *testing.T) {
		err := errors.New("some other error")
		if fsx.IsNotExist(err) {
			t.Fatal("expected false, got true")
		}
	})
}
