// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"io"
	"net/http"

	"github.com/rbmk-project/rbmk/pkg/x/netsim/dns"
)

// DNSHandler is an alias for [dns.Handler].
type DNSHandler = dns.Handler

// dnsDatabase is an alias for [dns.Database].
type dnsDatabase = dns.Database

// newDNSDatabase is an alias for [dns.NewDatabase].
var newDNSDatabase = dns.NewDatabase

// NewDNSHTTPHandler returns an [http.Handler] handling DNS-over-HTTPS.
func NewDNSHTTPHandler(dd dns.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		w.Header().Add("Content-Type", "application/dns-message")
		dd.Handle(w, rawQuery)
	})
}
