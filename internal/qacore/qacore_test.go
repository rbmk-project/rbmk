// SPDX-License-Identifier: GPL-3.0-or-later

package qacore_test

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/bassosimone/sud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchWwwExampleCom fetches https://www.example.com/ through the simulation
// and returns the response status code and body.
func fetchWwwExampleCom(t *testing.T) (int, string) {
	t.Helper()
	ctx := t.Context()

	// Resolve www.example.com
	addrs, err := simulation.LookupHost(ctx, "www.example.com")
	require.NoError(t, err)

	// Dial and TLS handshake
	endpoint := net.JoinHostPort(addrs[0], "443")
	conn, err := simulation.DialContext(ctx, "tcp", endpoint)
	require.NoError(t, err)
	defer conn.Close()

	tcfg := &tls.Config{
		ServerName: "www.example.com",
		RootCAs:    simulation.CertPool(),
	}
	tconn := tls.Client(conn, tcfg)
	defer tconn.Close()
	require.NoError(t, tconn.HandshakeContext(ctx))

	// HTTP round trip
	suse := sud.NewSingleUseDialer(tconn)
	txp := &http.Transport{DialTLSContext: suse.DialContext}
	clnt := &http.Client{Transport: txp}
	hr, err := clnt.Get("https://www.example.com/")
	require.NoError(t, err)
	defer hr.Body.Close()

	body, err := io.ReadAll(hr.Body)
	require.NoError(t, err)
	return hr.StatusCode, string(body)
}

func TestSetWwwExampleComHandler(t *testing.T) {
	const customBody = "custom response body"

	// Set a custom handler that returns a plain text response
	simulation.SetWwwExampleComHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(customBody))
	}))

	// Verify the custom handler is used
	status, body := fetchWwwExampleCom(t)
	assert.Equal(t, 200, status)
	assert.Equal(t, customBody, body)

	// Reset to default handler
	simulation.SetWwwExampleComHandler(nil)

	// Verify the default handler is restored
	status, body = fetchWwwExampleCom(t)
	assert.Equal(t, 200, status)
	assert.Equal(t, 605, len(body))
}
