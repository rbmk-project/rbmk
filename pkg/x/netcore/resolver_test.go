// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNetwork_maybeLookupEndpoint(t *testing.T) {
	t.Run("invalid endpoint format", func(t *testing.T) {
		nx := &Network{}
		_, err := nx.maybeLookupEndpoint(context.Background(), "invalid:endpoint:format")
		assert.Error(t, err)
	})

	t.Run("lookup error", func(t *testing.T) {
		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, expectedErr
			},
		}
		_, err := nx.maybeLookupEndpoint(context.Background(), "example.com:80")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("successful lookup", func(t *testing.T) {
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4", "5.6.7.8"}, nil
			},
		}
		endpoints, err := nx.maybeLookupEndpoint(context.Background(), "example.com:80")
		assert.NoError(t, err)
		assert.Equal(t, []string{"1.2.3.4:80", "5.6.7.8:80"}, endpoints)
	})
}

func TestNetwork_maybeLookupHost(t *testing.T) {
	t.Run("IP address short circuit", func(t *testing.T) {
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, errors.New("should not be called")
			},
		}
		addrs, err := nx.maybeLookupHost(context.Background(), "1.1.1.1")
		assert.NoError(t, err)
		assert.Equal(t, []string{"1.1.1.1"}, addrs)
	})

	t.Run("custom lookup success", func(t *testing.T) {
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4", "5.6.7.8"}, nil
			},
		}
		addrs, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.NoError(t, err)
		assert.Equal(t, []string{"1.2.3.4", "5.6.7.8"}, addrs)
	})

	t.Run("custom lookup error", func(t *testing.T) {
		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, expectedErr
			},
		}
		_, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("the timeout is correctly set and propagated", func(t *testing.T) {
		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				if _, ok := ctx.Deadline(); !ok {
					return nil, errors.New("missing context deadline")
				}
				select {
				case <-time.After(250 * time.Millisecond):
					return nil, errors.New("timeout not respected")
				case <-ctx.Done():
					return nil, expectedErr
				}
			},
			LookupHostTimeout: time.Microsecond,
		}
		_, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("system resolver error", func(t *testing.T) {
		nx := &Network{
			NewResolverOrSingleton: func() *net.Resolver {
				reso := &net.Resolver{}
				reso.PreferGo = true
				reso.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
					return nil, errors.New("mocked dial error")
				}
				return reso
			},
		}
		_, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.Error(t, err)
	})

	t.Run("logging behavior in case of success", func(t *testing.T) {
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
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return []string{"1.2.3.4", "5.6.7.8"}, nil
			},
		}

		addrs, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.NoError(t, err)
		assert.Equal(t, []string{"1.2.3.4", "5.6.7.8"}, addrs)

		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, logs, 2)

		// Verify lookupHostStart log
		var startLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[0]), &startLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":           "INFO",
			"msg":             "lookupHostStart",
			"dnsLookupDomain": "example.com",
			"t":               fixedTime.Format(time.RFC3339Nano),
		}, startLog)

		// Verify lookupHostDone log
		var doneLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[1]), &doneLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":            "INFO",
			"msg":              "lookupHostDone",
			"dnsLookupDomain":  "example.com",
			"dnsResolvedAddrs": []interface{}{"1.2.3.4", "5.6.7.8"},
			"err":              nil,
			"errClass":         "",
			"t0":               fixedTime.Format(time.RFC3339Nano),
			"t":                fixedTime.Format(time.RFC3339Nano),
		}, doneLog)
	})

	t.Run("logging behavior in case of error", func(t *testing.T) {
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

		expectedErr := errors.New("mocked lookup error")
		nx := &Network{
			Logger: logger,
			TimeNow: func() time.Time {
				return fixedTime
			},
			LookupHostFunc: func(ctx context.Context, domain string) ([]string, error) {
				return nil, expectedErr
			},
		}

		addrs, err := nx.maybeLookupHost(context.Background(), "example.com")
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, addrs)

		logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
		assert.Len(t, logs, 2)

		// Verify lookupHostStart log
		var startLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[0]), &startLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":           "INFO",
			"msg":             "lookupHostStart",
			"dnsLookupDomain": "example.com",
			"t":               fixedTime.Format(time.RFC3339Nano),
		}, startLog)

		// Verify lookupHostDone log
		var doneLog map[string]interface{}
		err = json.Unmarshal([]byte(logs[1]), &doneLog)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"level":            "INFO",
			"msg":              "lookupHostDone",
			"dnsLookupDomain":  "example.com",
			"dnsResolvedAddrs": nil,
			"err":              expectedErr.Error(),
			"errClass":         "EGENERIC",
			"t0":               fixedTime.Format(time.RFC3339Nano),
			"t":                fixedTime.Format(time.RFC3339Nano),
		}, doneLog)
	})
}
