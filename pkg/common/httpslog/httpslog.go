// SPDX-License-Identifier: GPL-3.0-or-later

// Package httpslog implements structured logging for HTTP clients.
package httpslog

import (
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/errclass"
)

// MaybeLogRoundTripStart logs the start of a round trip if the
// given logger is not nil, otherwise it does nothing.
func MaybeLogRoundTripStart(
	logger *slog.Logger,
	localAddr netip.AddrPort,
	protocol string,
	remoteAddr netip.AddrPort,
	req *http.Request,
	t0 time.Time,
) {
	if logger != nil {
		logger.InfoContext(
			req.Context(),
			"httpRoundTripStart",
			slog.String("httpMethod", req.Method),
			slog.String("httpUrl", req.URL.String()),
			slog.Any("httpRequestHeaders", req.Header),
			slog.String("localAddr", localAddr.String()),
			slog.String("protocol", protocol),
			slog.String("remoteAddr", remoteAddr.String()),
			slog.Time("t", t0),
		)
	}
}

// MaybeLogRoundTripDone logs the end of a round trip if the given
// logger is not nil, otherwise it does nothing.
func MaybeLogRoundTripDone(
	logger *slog.Logger,
	localAddr netip.AddrPort,
	protocol string,
	remoteAddr netip.AddrPort,
	req *http.Request,
	resp *http.Response,
	err error,
	t0 time.Time,
	t time.Time,
) {
	if logger != nil {
		if err != nil {
			logger.InfoContext(
				req.Context(),
				"httpRoundTripDone",
				slog.Any("err", err),
				slog.Any("errClass", errclass.New(err)),
				slog.String("httpMethod", req.Method),
				slog.String("httpUrl", req.URL.String()),
				slog.Any("httpRequestHeaders", req.Header),
				slog.String("localAddr", localAddr.String()),
				slog.String("protocol", protocol),
				slog.String("remoteAddr", remoteAddr.String()),
				slog.Time("t0", t0),
				slog.Time("t", t),
			)
			return
		}
		logger.InfoContext(
			req.Context(),
			"httpRoundTripDone",
			slog.String("httpMethod", req.Method),
			slog.String("httpUrl", req.URL.String()),
			slog.Any("httpRequestHeaders", req.Header),
			slog.Int("httpResponseStatusCode", resp.StatusCode),
			slog.Any("httpResponseHeaders", resp.Header),
			slog.String("localAddr", localAddr.String()),
			slog.String("protocol", protocol),
			slog.String("remoteAddr", remoteAddr.String()),
			slog.Time("t", t0),
		)
	}
}
