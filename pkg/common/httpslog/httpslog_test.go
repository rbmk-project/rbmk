// SPDX-License-Identifier: GPL-3.0-or-later

package httpslog

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaybeLogRoundTripStart(t *testing.T) {
	tests := []struct {
		name       string
		newLogger  func(w io.Writer) *slog.Logger
		expectTime time.Time
		expectLog  string
	}{
		{
			name: "Logger set",
			newLogger: func(w io.Writer) *slog.Logger {
				return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
					Level: slog.LevelDebug,
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if attr.Key == slog.TimeKey {
							return slog.Attr{}
						}
						return attr
					},
				}))
			},
			expectTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectLog: `{"level":"INFO","msg":"httpRoundTripStart","httpMethod":"GET",` +
				`"httpUrl":"https://example.com","httpRequestHeaders":{},"localAddr":"127.0.0.1:0",` +
				`"protocol":"tcp","remoteAddr":"93.184.216.34:443","t":"2020-01-01T00:00:00Z"}` + "\n",
		},
		{
			name:       "Logger not set",
			newLogger:  func(w io.Writer) *slog.Logger { return nil },
			expectTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectLog:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			logger := tt.newLogger(&out)

			req, err := http.NewRequest("GET", "https://example.com", nil)
			assert.NoError(t, err)

			localAddr := netip.MustParseAddrPort("127.0.0.1:0")
			remoteAddr := netip.MustParseAddrPort("93.184.216.34:443")

			MaybeLogRoundTripStart(
				logger,
				localAddr,
				"tcp",
				remoteAddr,
				req,
				tt.expectTime,
			)

			actualLog := out.String()
			assert.Equal(t, tt.expectLog, actualLog)
		})
	}
}

func TestMaybeLogRoundTripDone(t *testing.T) {
	tests := []struct {
		name       string
		newLogger  func(w io.Writer) *slog.Logger
		withError  bool
		expectTime time.Time
		expectLog  string
	}{
		{
			name: "Logger set with success",
			newLogger: func(w io.Writer) *slog.Logger {
				return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
					Level: slog.LevelDebug,
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if attr.Key == slog.TimeKey {
							return slog.Attr{}
						}
						return attr
					},
				}))
			},
			withError:  false,
			expectTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectLog: `{"level":"INFO","msg":"httpRoundTripDone","httpMethod":"GET",` +
				`"httpUrl":"https://example.com","httpRequestHeaders":{},` +
				`"httpResponseStatusCode":200,"httpResponseHeaders":{},` +
				`"localAddr":"127.0.0.1:0","protocol":"tcp","remoteAddr":"93.184.216.34:443",` +
				`"t":"2020-01-01T00:00:00Z"}` + "\n",
		},
		{
			name: "Logger set with error",
			newLogger: func(w io.Writer) *slog.Logger {
				return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
					Level: slog.LevelDebug,
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if attr.Key == slog.TimeKey {
							return slog.Attr{}
						}
						return attr
					},
				}))
			},
			withError:  true,
			expectTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectLog: `{"level":"INFO","msg":"httpRoundTripDone","err":"assert.AnError general error for testing",` +
				`"errClass":"EGENERIC","httpMethod":"GET","httpUrl":"https://example.com",` +
				`"httpRequestHeaders":{},"localAddr":"127.0.0.1:0","protocol":"tcp",` +
				`"remoteAddr":"93.184.216.34:443","t0":"2020-01-01T00:00:00Z","t":"2020-01-01T00:00:00Z"}` + "\n",
		},
		{
			name:       "Logger not set",
			newLogger:  func(w io.Writer) *slog.Logger { return nil },
			withError:  false,
			expectTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectLog:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			logger := tt.newLogger(&out)

			req, err := http.NewRequest("GET", "https://example.com", nil)
			assert.NoError(t, err)

			var resp *http.Response
			var roundTripErr error

			if !tt.withError {
				resp = &http.Response{
					StatusCode: 200,
					Header:     make(http.Header),
				}
			} else {
				roundTripErr = assert.AnError
			}

			localAddr := netip.MustParseAddrPort("127.0.0.1:0")
			remoteAddr := netip.MustParseAddrPort("93.184.216.34:443")

			MaybeLogRoundTripDone(
				logger,
				localAddr,
				"tcp",
				remoteAddr,
				req,
				resp,
				roundTripErr,
				tt.expectTime,
				tt.expectTime,
			)

			actualLog := out.String()
			assert.Equal(t, tt.expectLog, actualLog)

			// Verify JSON is valid when there's output
			if actualLog != "" {
				var jsonMap map[string]interface{}
				err := json.Unmarshal([]byte(actualLog), &jsonMap)
				assert.NoError(t, err)
			}
		})
	}
}
