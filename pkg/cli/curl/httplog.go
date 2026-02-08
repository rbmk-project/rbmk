// SPDX-License-Identifier: GPL-3.0-or-later

package curl

import (
	"net/http"

	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
)

// httpLogTransport is an [http.RoundTripper] that emits logs
//
// Construct using [newHTTPLogTransport].
type httpLogTransport struct {
	pool *closepool.Pool
	netx *netcore.Network
}

// newHTTPLogTransport constructs a new [*httpLogTransport].
func newHTTPLogTransport(netx *netcore.Network, pool *closepool.Pool) *httpLogTransport {
	return &httpLogTransport{pool: pool, netx: netx}
}

var _ http.RoundTripper = &httpLogTransport{}

// RoundTrip implements [http.RoundTripper].
func (txp *httpLogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	conn, err := txp.netx.DialHTTP(req)
	if err != nil {
		return nil, err
	}
	txp.pool.Add(conn)
	return conn.RoundTrip(req)
}
