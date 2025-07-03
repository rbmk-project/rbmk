// SPDX-License-Identifier: GPL-3.0-or-later

package curl

import (
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/rbmk-project/rbmk/pkg/common/httpconntrace"
	"github.com/rbmk-project/rbmk/pkg/common/httpslog"
)

// httpDoAndLog performs the request and emits structured logs.
//
// We *assume* that the client does not follow redirects, which means
// the connection we observe is the only one being used.
//
// Note: the slogger *may* be nil.
//
// We use [httpconntrace] to extract the local and remote addresses
// emitted as part of the structured log events.
func httpDoAndLog(
	client *http.Client,
	slogger *slog.Logger,
	req *http.Request,
) (*http.Response, error) {
	// possibly emit a structured log event before performing the request
	t0 := time.Now()
	httpslog.MaybeLogRoundTripStart(
		slogger,
		netip.MustParseAddrPort("[::]:0"), // not known yet
		"tcp",
		netip.MustParseAddrPort("[::]:0"), // not known yet
		req,
		t0,
	)

	// perform the request
	resp, epnts, err := httpconntrace.Do(client, req)

	// possibly emit a structured log event after performing the request
	t := time.Now()
	httpslog.MaybeLogRoundTripDone(
		slogger,
		epnts.LocalAddr,
		"tcp",
		epnts.RemoteAddr,
		req,
		resp,
		err,
		t0,
		t,
	)

	// Forward the results to the caller.
	return resp, err
}
