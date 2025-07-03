// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/rbmk-project/common/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_connLocalAddr(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		addr := connLocalAddr(nil)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("nil local address", func(t *testing.T) {
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr { return nil },
		}
		addr := connLocalAddr(conn)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("valid address", func(t *testing.T) {
		expectedAddr := &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 1234,
		}
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr { return expectedAddr },
		}
		addr := connLocalAddr(conn)
		assert.Equal(t, expectedAddr, addr)
	})
}

func Test_connRemoteAddr(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		addr := connRemoteAddr(nil)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("nil remote address", func(t *testing.T) {
		conn := &mocks.Conn{
			MockRemoteAddr: func() net.Addr { return nil },
		}
		addr := connRemoteAddr(conn)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("valid address", func(t *testing.T) {
		expectedAddr := &net.TCPAddr{
			IP:   net.ParseIP("1.1.1.1"),
			Port: 443,
		}
		conn := &mocks.Conn{
			MockRemoteAddr: func() net.Addr { return expectedAddr },
		}
		addr := connRemoteAddr(conn)
		assert.Equal(t, expectedAddr, addr)
	})
}

func TestNetwork_maybeWrapConn(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		nx := &Network{}
		assert.Nil(t, nx.maybeWrapConn(context.Background(), nil))
	})

	t.Run("no logger configured", func(t *testing.T) {
		nx := &Network{}
		conn := &mocks.Conn{}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.Equal(t, conn, wrapped) // should return unwrapped
	})

	t.Run("no wrapper configured", func(t *testing.T) {
		nx := &Network{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}
		conn := &mocks.Conn{}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.Equal(t, conn, wrapped) // should return unwrapped
	})

	t.Run("full wrapping", func(t *testing.T) {
		nx := &Network{
			Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			WrapConn: WrapConn,
		}
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
			},
		}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.NotEqual(t, conn, wrapped) // should return wrapped
		inner, ok := wrapped.(*connWrapper)
		assert.True(t, ok)
		assert.Equal(t, inner.netx, nx)
	})
}

func TestWrapConn(t *testing.T) {
	t.Run("correctly initializes wrapper", func(t *testing.T) {
		nx := &Network{
			Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			WrapConn: WrapConn,
		}
		mock := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
			},
		}

		conn := WrapConn(context.Background(), nx, mock)
		wrapped, ok := conn.(*connWrapper)
		assert.True(t, ok)
		assert.Equal(t, nx, wrapped.netx)
	})

	t.Run("handles nil addresses gracefully", func(t *testing.T) {
		nx := &Network{}
		mock := &mocks.Conn{
			MockLocalAddr:  func() net.Addr { return nil },
			MockRemoteAddr: func() net.Addr { return nil },
		}

		conn := WrapConn(context.Background(), nx, mock)
		wrapped, ok := conn.(*connWrapper)
		assert.True(t, ok)
		assert.Equal(t, "", wrapped.laddr)
		assert.Equal(t, "", wrapped.raddr)
	})
}

func Test_connWrapper(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		// Helper function to create a standard test environment
		setup := func() (*bytes.Buffer, *mocks.Conn, *connWrapper, time.Time) {
			var buf bytes.Buffer
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			timeNow := func() time.Time {
				return fixedTime
			}

			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			mock := &mocks.Conn{
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{Logger: logger, TimeNow: timeNow},
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			return &buf, mock, wrapper, fixedTime
		}

		t.Run("successful close", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock behavior
			mock.MockClose = func() error {
				return nil
			}

			// Perform the close operation
			err := wrapper.Close()
			assert.NoError(t, err)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify closeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeStart",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify closeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeDone",
				"err":        nil,
				"errClass":   "",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t0":         fixedTime.Format(time.RFC3339Nano),
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("error on close", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock to return error
			expectedErr := errors.New("mocked close error")
			mock.MockClose = func() error {
				return expectedErr
			}

			// Perform the close operation
			err := wrapper.Close()
			assert.ErrorIs(t, err, expectedErr)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify closeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeStart",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify closeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeDone",
				"err":        expectedErr.Error(),
				"errClass":   "EGENERIC",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t0":         fixedTime.Format(time.RFC3339Nano),
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("idempotent close", func(t *testing.T) {
			buf, mock, wrapper, _ := setup()

			closeCount := 0
			mock.MockClose = func() error {
				closeCount++
				return nil
			}

			// Close multiple times
			err1 := wrapper.Close()
			err2 := wrapper.Close()
			err3 := wrapper.Close()

			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.NoError(t, err3)
			assert.Equal(t, 1, closeCount, "Close should only be called once")

			// Verify we only logged one close operation
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2, "Should only have one pair of start/done logs")
		})

		t.Run("no logger configured", func(t *testing.T) {
			// Implementation note: this covers the case where you use WrapConn to create
			// a connWrapper where the underlying netx has no configured logger.

			mock := &mocks.Conn{
				MockClose: func() error {
					return nil
				},
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{}, // no logger configured
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			err := wrapper.Close()
			assert.NoError(t, err)
		})
	})

	t.Run("Read", func(t *testing.T) {
		// Helper function to create a standard test environment
		setup := func() (*bytes.Buffer, *mocks.Conn, *connWrapper, time.Time) {
			var buf bytes.Buffer
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			timeNow := func() time.Time {
				return fixedTime
			}

			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			mock := &mocks.Conn{
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{Logger: logger, TimeNow: timeNow},
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			return &buf, mock, wrapper, fixedTime
		}

		t.Run("successful read", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock behavior
			expectedData := []byte("hello world")
			mock.MockRead = func(b []byte) (int, error) {
				copy(b, expectedData)
				return len(expectedData), nil
			}

			// Perform the read operation
			readBuf := make([]byte, 1024)
			n, err := wrapper.Read(readBuf)
			assert.NoError(t, err)
			assert.Equal(t, len(expectedData), n)
			assert.Equal(t, expectedData, readBuf[:n])

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify readStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "readStart",
				"ioBufferSize": float64(1024),
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify readDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "readDone",
				"ioBytesCount": float64(len(expectedData)),
				"err":          nil,
				"errClass":     "",
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t0":           fixedTime.Format(time.RFC3339Nano),
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("read with error", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock to return error
			expectedErr := errors.New("mocked read error")
			mock.MockRead = func(b []byte) (int, error) {
				return 0, expectedErr
			}

			// Perform the read operation
			readBuf := make([]byte, 1024)
			n, err := wrapper.Read(readBuf)
			assert.ErrorIs(t, err, expectedErr)
			assert.Zero(t, n)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify readStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "readStart",
				"ioBufferSize": float64(1024),
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify readDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "readDone",
				"ioBytesCount": float64(0),
				"err":          expectedErr.Error(),
				"errClass":     "EGENERIC",
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t0":           fixedTime.Format(time.RFC3339Nano),
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("no logger configured", func(t *testing.T) {
			mock := &mocks.Conn{
				MockRead: func(b []byte) (int, error) {
					return len("test"), nil
				},
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{}, // no logger configured
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			readBuf := make([]byte, 1024)
			n, err := wrapper.Read(readBuf)
			assert.NoError(t, err)
			assert.Equal(t, len("test"), n)
		})
	})

	t.Run("Write", func(t *testing.T) {
		// Helper function to create a standard test environment
		setup := func() (*bytes.Buffer, *mocks.Conn, *connWrapper, time.Time) {
			var buf bytes.Buffer
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			timeNow := func() time.Time {
				return fixedTime
			}

			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			mock := &mocks.Conn{
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{Logger: logger, TimeNow: timeNow},
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			return &buf, mock, wrapper, fixedTime
		}

		t.Run("successful write", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock behavior
			data := []byte("hello world")
			mock.MockWrite = func(b []byte) (int, error) {
				return len(b), nil
			}

			// Perform the write operation
			n, err := wrapper.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify writeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "writeStart",
				"ioBufferSize": float64(len(data)),
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify writeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "writeDone",
				"ioBytesCount": float64(len(data)),
				"err":          nil,
				"errClass":     "",
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t0":           fixedTime.Format(time.RFC3339Nano),
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("write with error", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock to return error
			expectedErr := errors.New("mocked write error")
			mock.MockWrite = func(b []byte) (int, error) {
				return 0, expectedErr
			}

			// Perform the write operation
			data := []byte("hello world")
			n, err := wrapper.Write(data)
			assert.ErrorIs(t, err, expectedErr)
			assert.Zero(t, n)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify writeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "writeStart",
				"ioBufferSize": float64(len(data)),
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify writeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":        "INFO",
				"msg":          "writeDone",
				"ioBytesCount": float64(0),
				"err":          expectedErr.Error(),
				"errClass":     "EGENERIC",
				"localAddr":    "127.0.0.1:1234",
				"protocol":     "tcp",
				"remoteAddr":   "1.1.1.1:443",
				"t0":           fixedTime.Format(time.RFC3339Nano),
				"t":            fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("no logger configured", func(t *testing.T) {
			mock := &mocks.Conn{
				MockWrite: func(b []byte) (int, error) {
					return len(b), nil
				},
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{}, // no logger configured
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			data := []byte("test")
			n, err := wrapper.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)
		})
	})

	t.Run("LocalAddr", func(t *testing.T) {
		// Create a mock connection with a specific local address
		expectedAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
		mockConn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return expectedAddr
			},
		}

		// Create a wrapper around the mock
		wrapper := &connWrapper{
			conn: mockConn,
			// Other fields aren't relevant for this test
		}

		// Test the LocalAddr method
		addr := wrapper.LocalAddr()

		// Verify it returns the expected address
		assert.Equal(t, expectedAddr, addr)
	})

	t.Run("RemoteAddr", func(t *testing.T) {
		// Create a mock connection with a specific remote address
		expectedAddr := &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
		mockConn := &mocks.Conn{
			MockRemoteAddr: func() net.Addr {
				return expectedAddr
			},
		}

		// Create a wrapper around the mock
		wrapper := &connWrapper{
			conn: mockConn,
			// Other fields aren't relevant for this test
		}

		// Test the RemoteAddr method
		addr := wrapper.RemoteAddr()

		// Verify it returns the expected address
		assert.Equal(t, expectedAddr, addr)
	})

	t.Run("SetDeadline", func(t *testing.T) {
		// Track if SetDeadline was called with the correct value
		var calledWithTime time.Time

		mockConn := &mocks.Conn{
			MockSetDeadline: func(t time.Time) error {
				calledWithTime = t
				return nil
			},
		}

		// Create a wrapper around the mock
		wrapper := &connWrapper{
			conn: mockConn,
		}

		// Call the method with a specific time
		expectedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		err := wrapper.SetDeadline(expectedTime)

		// Verify the method call was passed through to the underlying conn
		assert.NoError(t, err)
		assert.Equal(t, expectedTime, calledWithTime)

		// Test error propagation
		expectedErr := errors.New("deadline error")
		mockConn.MockSetDeadline = func(t time.Time) error {
			return expectedErr
		}

		err = wrapper.SetDeadline(time.Now())
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("SetReadDeadline", func(t *testing.T) {
		// Track if SetReadDeadline was called with the correct value
		var calledWithTime time.Time

		mockConn := &mocks.Conn{
			MockSetReadDeadline: func(t time.Time) error {
				calledWithTime = t
				return nil
			},
		}

		// Create a wrapper around the mock
		wrapper := &connWrapper{
			conn: mockConn,
		}

		// Call the method with a specific time
		expectedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		err := wrapper.SetReadDeadline(expectedTime)

		// Verify the method call was passed through to the underlying conn
		assert.NoError(t, err)
		assert.Equal(t, expectedTime, calledWithTime)

		// Test error propagation
		expectedErr := errors.New("read deadline error")
		mockConn.MockSetReadDeadline = func(t time.Time) error {
			return expectedErr
		}

		err = wrapper.SetReadDeadline(time.Now())
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("SetWriteDeadline", func(t *testing.T) {
		// Track if SetWriteDeadline was called with the correct value
		var calledWithTime time.Time

		mockConn := &mocks.Conn{
			MockSetWriteDeadline: func(t time.Time) error {
				calledWithTime = t
				return nil
			},
		}

		// Create a wrapper around the mock
		wrapper := &connWrapper{
			conn: mockConn,
		}

		// Call the method with a specific time
		expectedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		err := wrapper.SetWriteDeadline(expectedTime)

		// Verify the method call was passed through to the underlying conn
		assert.NoError(t, err)
		assert.Equal(t, expectedTime, calledWithTime)

		// Test error propagation
		expectedErr := errors.New("write deadline error")
		mockConn.MockSetWriteDeadline = func(t time.Time) error {
			return expectedErr
		}

		err = wrapper.SetWriteDeadline(time.Now())
		assert.ErrorIs(t, err, expectedErr)
	})
}
