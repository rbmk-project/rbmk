// SPDX-License-Identifier: GPL-3.0-or-later

package closepool_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/closepool"
)

// mockCloser implements io.Closer for testing
type mockCloser struct {
	closed atomic.Int64
	err    error
}

// t0 is the time when we started running
var t0 = time.Now()

func (m *mockCloser) Close() error {
	m.closed.Add(int64(time.Since(t0)))
	return m.err
}

func TestCloserFunc(t *testing.T) {
	var closed bool
	pool := &closepool.Pool{}
	pool.Add(closepool.CloserFunc(func() error {
		closed = true
		return nil
	}))
	pool.Close()
	if !closed {
		t.Error("expected closer to be closed")
	}
}

func TestPool(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		pool := closepool.Pool{}
		m1 := &mockCloser{}
		m2 := &mockCloser{}

		pool.Add(m1)
		pool.Add(m2)

		err := pool.Close()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if m1.closed.Load() <= 0 {
			t.Error("first closer was not closed")
		}
		if m2.closed.Load() <= 0 {
			t.Error("second closer was not closed")
		}
	})

	t.Run("close order", func(t *testing.T) {
		pool := closepool.Pool{}

		m1 := &mockCloser{
			err: nil,
		}
		m2 := &mockCloser{
			err: nil,
		}

		pool.Add(m1) // Added first
		pool.Add(m2) // Added second

		// Should close in reverse order
		err := pool.Close()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if m1.closed.Load() <= m2.closed.Load() {
			t.Error("expected m1 to be closed after m2")
		}
	})

	t.Run("error handling", func(t *testing.T) {
		pool := closepool.Pool{}
		expectedErr1 := errors.New("close error #1")
		expectedErr2 := errors.New("close error #2")

		m1 := &mockCloser{err: expectedErr1}
		m2 := &mockCloser{err: expectedErr2}

		pool.Add(m1)
		pool.Add(m2)

		err := pool.Close()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		t.Log(err)
		if errors.Join(expectedErr2, expectedErr1).Error() != err.Error() {
			t.Errorf("expected error to include both errors, got %v", err)
		}
	})

	t.Run("concurrent usage", func(t *testing.T) {
		pool := closepool.Pool{}
		done := make(chan struct{})

		// Concurrently add closers
		go func() {
			for i := 0; i < 100; i++ {
				pool.Add(&mockCloser{})
			}
			close(done)
		}()

		// Add more closers from main goroutine
		for i := 0; i < 100; i++ {
			pool.Add(&mockCloser{})
		}

		<-done // Wait for goroutine to finish

		err := pool.Close()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
