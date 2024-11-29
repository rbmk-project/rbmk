// SPDX-License-Identifier: GPL-3.0-or-later

package curl

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/rbmk-project/x/errclass"
)

// TODO(bassosimone): we should probably share this
// functionality with the `dig` task as well.

// httpLogTransport is an http.RoundTripper that logs.
type httpLogTransport struct {
	// Logger is the optional Logger to use.
	Logger *slog.Logger

	// RoundTripper is the mandatory http.RoundTripper to use.
	RoundTripper http.RoundTripper

	// TimeNow is the optional time function to use.
	TimeNow func() time.Time
}

// RoundTrip implements [http.RoundTripper].
func (txp *httpLogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO(bassosimone): adapt support for obtaining the
	// connection information from the dnscore package

	// possibly emit a structured log event before the round trip
	t0 := txp.timeNow()
	if txp.Logger != nil {
		txp.Logger.InfoContext(
			req.Context(),
			"httpRoundTripStart",
			slog.String("httpMethod", req.Method),
			slog.String("httpUrl", req.URL.String()),
			slog.Any("httpRequestHeaders", req.Header),
			slog.Time("t", t0),
		)
	}

	// perform the actual round trip
	resp, err := txp.RoundTripper.RoundTrip(req)
	t := txp.timeNow()

	// handle the case of failure
	if err != nil {
		if txp.Logger != nil {
			txp.Logger.InfoContext(
				req.Context(),
				"httpRoundTripDone",
				slog.String("errClass", errclass.New(err)),
				slog.Any("err", err),
				slog.String("httpMethod", req.Method),
				slog.String("httpUrl", req.URL.String()),
				slog.Any("httpRequestHeaders", req.Header),
				slog.Time("t0", t0),
				slog.Time("t", t),
			)
		}
		return nil, err
	}

	// handle the case of success
	txp.Logger.InfoContext(
		req.Context(),
		"httpRoundTripDone",
		slog.String("httpMethod", req.Method),
		slog.String("httpUrl", req.URL.String()),
		slog.Any("httpRequestHeaders", req.Header),
		slog.Int("httpResponseStatusCode", resp.StatusCode),
		slog.Any("httpResponseHeaders", resp.Header),
		slog.Time("t0", t0),
		slog.Time("t", t),
	)
	return resp, nil
}

// timeNow returns the current time.
func (txp *httpLogTransport) timeNow() time.Time {
	if txp.TimeNow != nil {
		return txp.TimeNow()
	}
	return time.Now()
}
