// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rbmk-project/common/mocks"
	"github.com/rbmk-project/common/runtimex"
	"github.com/stretchr/testify/assert"
)

func TestNetwork_DialContext(t *testing.T) {
	t.Run("lookup failure", func(t *testing.T) {
		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, expectedErr
			},
		}
		conn, err := nx.DialContext(context.Background(), "tcp", "example.com:80")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("dial failure", func(t *testing.T) {
		expectedErr := errors.New("mocked dial error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, expectedErr
			},
		}
		conn, err := nx.DialContext(context.Background(), "tcp", "example.com:80")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)
	})

	t.Run("successful dial", func(t *testing.T) {
		mockConn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 80}
			},
		}
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4"}, nil
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
		}
		conn, err := nx.DialContext(context.Background(), "tcp", "example.com:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)
	})
}

func TestNetwork_sequentialDial(t *testing.T) {
	t.Run("empty endpoints list", func(t *testing.T) {
		nx := &Network{}
		conn, err := nx.sequentialDial(context.Background(), "tcp", nx.dialLog)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("all endpoints fail", func(t *testing.T) {
		expectedErr1 := errors.New("error 1")
		expectedErr2 := errors.New("error 2")
		dialAttempts := 0
		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialAttempts++
				if address == "1.1.1.1:80" {
					return nil, expectedErr1
				}
				return nil, expectedErr2
			},
		}
		conn, err := nx.sequentialDial(context.Background(), "tcp", nx.dialLog, "1.1.1.1:80", "2.2.2.2:80")
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Equal(t, 2, dialAttempts)
		assert.ErrorIs(t, err, expectedErr1)
		assert.ErrorIs(t, err, expectedErr2)
	})

	t.Run("first endpoint succeeds", func(t *testing.T) {
		mockConn := &mocks.Conn{}
		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
		}
		conn, err := nx.sequentialDial(context.Background(), "tcp", nx.dialLog, "1.1.1.1:80", "2.2.2.2:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)
	})

	t.Run("second endpoint succeeds", func(t *testing.T) {
		mockConn := &mocks.Conn{}
		expectedErr := errors.New("first endpoint fails")
		dialAttempts := 0
		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialAttempts++
				if dialAttempts == 1 {
					return nil, expectedErr
				}
				return mockConn, nil
			},
		}
		conn, err := nx.sequentialDial(context.Background(), "tcp", nx.dialLog, "1.1.1.1:80", "2.2.2.2:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)
		assert.Equal(t, 2, dialAttempts)
	})
}

func TestNetwork_dialLog(t *testing.T) {
	t.Run("successful dial with logging", func(t *testing.T) {
		var buf bytes.Buffer
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}))

		mockConn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 80}
			},
		}

		nx := &Network{
			Logger: logger,
			TimeNow: func() time.Time {
				return fixedTime
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				deadline, hasDeadline := ctx.Deadline()
				assert.False(t, hasDeadline, "context should not have deadline")
				assert.Equal(t, time.Time{}, deadline)
				return mockConn, nil
			},
		}

		conn, err := nx.dialLog(context.Background(), "tcp", "1.1.1.1:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)

		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, logs, 2)

		// Verify connectStart log
		var startLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[0]), &startLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectStart",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, startLog)

		// Verify connectDone log
		var doneLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[1]), &doneLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectDone",
			"err":        nil,
			"errClass":   "",
			"localAddr":  "127.0.0.1:1234",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t0":         fixedTime.Format(time.RFC3339Nano),
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, doneLog)
	})

	t.Run("dial failure with logging", func(t *testing.T) {
		var buf bytes.Buffer
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}))

		expectedErr := errors.New("mocked dial error")
		nx := &Network{
			Logger: logger,
			TimeNow: func() time.Time {
				return fixedTime
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, expectedErr
			},
		}

		conn, err := nx.dialLog(context.Background(), "tcp", "1.1.1.1:80")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, conn)

		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, logs, 2)

		// Verify connectStart log
		var startLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[0]), &startLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectStart",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, startLog)

		// Verify connectDone log
		var doneLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[1]), &doneLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectDone",
			"err":        expectedErr.Error(),
			"errClass":   "EGENERIC",
			"localAddr":  "",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t0":         fixedTime.Format(time.RFC3339Nano),
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, doneLog)
	})

	t.Run("successful dial without logging", func(t *testing.T) {
		mockConn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 80}
			},
		}

		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
		}

		conn, err := nx.dialLog(context.Background(), "tcp", "1.1.1.1:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)
	})

	t.Run("dial timeout exceeded", func(t *testing.T) {
		var buf bytes.Buffer
		fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}))

		nx := &Network{
			Logger: logger,
			TimeNow: func() time.Time {
				return fixedTime
			},
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				_, hasDeadline := ctx.Deadline()
				assert.True(t, hasDeadline, "context should have deadline")

				// Simulate slow operation that should be interrupted by timeout
				select {
				case <-time.After(250 * time.Millisecond):
					return nil, errors.New("timeout not respected")
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
			DialContextTimeout: time.Microsecond,
		}

		conn, err := nx.dialLog(context.Background(), "tcp", "1.1.1.1:80")
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, conn)

		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, logs, 2)

		// Verify connectStart log
		var startLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[0]), &startLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectStart",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, startLog)

		// Verify connectDone log
		var doneLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[1]), &doneLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":      "INFO",
			"msg":        "connectDone",
			"err":        context.DeadlineExceeded.Error(),
			"errClass":   "ETIMEDOUT",
			"localAddr":  "",
			"protocol":   "tcp",
			"remoteAddr": "1.1.1.1:80",
			"t0":         fixedTime.Format(time.RFC3339Nano),
			"t":          fixedTime.Format(time.RFC3339Nano),
		}, doneLog)
	})
}

func TestNetwork_dialNet(t *testing.T) {
	t.Run("using custom DialContextFunc", func(t *testing.T) {
		mockConn := &mocks.Conn{}
		nx := &Network{
			DialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return mockConn, nil
			},
		}
		conn, err := nx.dialNet(context.Background(), "tcp", "1.1.1.1:80")
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)
	})

	t.Run("using custom NewDialerOrSingleton", func(t *testing.T) {
		// create a server using localhost to test against
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		nx := &Network{
			NewDialerOrSingleton: func() *net.Dialer {
				d := &net.Dialer{}
				return d
			},
		}
		parsed := runtimex.Try1(url.Parse(server.URL))
		conn, err := nx.dialNet(context.Background(), "tcp", parsed.Host)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		conn.Close()
	})

	t.Run("using the default dialer", func(t *testing.T) {
		// create a server using localhost to test against
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		nx := &Network{}
		parsed := runtimex.Try1(url.Parse(server.URL))
		conn, err := nx.dialNet(context.Background(), "tcp", parsed.Host)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		conn.Close()
	})
}
