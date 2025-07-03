//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/runtimex/runtimex_test.go
//

package runtimex_test

import (
	"errors"
	"testing"

	"github.com/rbmk-project/rbmk/pkg/common/runtimex"
)

func TestPanicOnError(t *testing.T) {
	badfunc := func(in error) (out error) {
		defer func() {
			out = recover().(error)
		}()
		runtimex.PanicOnError(in, "we expect this assertion to fail")
		return
	}

	t.Run("error is nil", func(t *testing.T) {
		runtimex.PanicOnError(nil, "this assertion should not fail")
	})

	t.Run("error is not nil", func(t *testing.T) {
		expected := errors.New("mocked error")
		if !errors.Is(badfunc(expected), expected) {
			t.Fatal("not the error we expected")
		}
	})
}

func TestAssert(t *testing.T) {
	badfunc := func(in bool, message string) (out error) {
		defer func() {
			out = recover().(error)
		}()
		runtimex.Assert(in, message)
		return
	}

	t.Run("assertion is true", func(t *testing.T) {
		runtimex.Assert(true, "this assertion should not fail")
	})

	t.Run("assertion is false", func(t *testing.T) {
		message := "mocked error"
		err := badfunc(false, message)
		if err == nil || err.Error() != message {
			t.Fatal("not the error we expected", err)
		}
	})
}

func TestTry(t *testing.T) {
	t.Run("Try0", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			runtimex.Try0(nil)
		})

		t.Run("on failure", func(t *testing.T) {
			expected := errors.New("mocked error")
			var got error
			func() {
				defer func() {
					if r := recover(); r != nil {
						got = r.(error)
					}
				}()
				runtimex.Try0(expected)
			}()
			if !errors.Is(got, expected) {
				t.Fatal("unexpected error")
			}
		})
	})

	t.Run("Try1", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			v1 := runtimex.Try1(17, nil)
			if v1 != 17 {
				t.Fatal("unexpected value")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			expected := errors.New("mocked error")
			var got error
			func() {
				defer func() {
					if r := recover(); r != nil {
						got = r.(error)
					}
				}()
				runtimex.Try1(17, expected)
			}()
			if !errors.Is(got, expected) {
				t.Fatal("unexpected error")
			}
		})
	})

	t.Run("Try2", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			v1, v2 := runtimex.Try2(17, true, nil)
			if v1 != 17 || !v2 {
				t.Fatal("unexpected value")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			expected := errors.New("mocked error")
			var got error
			func() {
				defer func() {
					if r := recover(); r != nil {
						got = r.(error)
					}
				}()
				runtimex.Try2(17, true, expected)
			}()
			if !errors.Is(got, expected) {
				t.Fatal("unexpected error")
			}
		})
	})

	t.Run("Try3", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			v1, v2, v3 := runtimex.Try3(17, true, 44.0, nil)
			if v1 != 17 || !v2 || v3 != 44.0 {
				t.Fatal("unexpected value")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			expected := errors.New("mocked error")
			var got error
			func() {
				defer func() {
					if r := recover(); r != nil {
						got = r.(error)
					}
				}()
				runtimex.Try3(17, true, 44.0, expected)
			}()
			if !errors.Is(got, expected) {
				t.Fatal("unexpected error")
			}
		})
	})
}
