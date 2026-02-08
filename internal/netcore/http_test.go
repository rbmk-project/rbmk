// SPDX-License-Identifier: GPL-3.0-or-later

package netcore_test

import (
	"net/http"
	"testing"

	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify that DialHTTP returns error for an unsupported scheme.
func TestDialHTTP_UnsupportedScheme(t *testing.T) {
	nx := netcore.NewNetwork()
	req, err := http.NewRequestWithContext(t.Context(), "GET", "ftp://www.example.com/", nil)
	require.NoError(t, err)
	_, err = nx.DialHTTP(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme")
}

// Verify that HTTPS round trip works via the simulation.
func TestDialHTTP_HTTPS(t *testing.T) {
	nx := netcore.NewNetwork()
	req, err := http.NewRequestWithContext(t.Context(), "GET", "https://www.example.com/", nil)
	require.NoError(t, err)
	conn, err := nx.DialHTTP(req)
	require.NoError(t, err)
	defer conn.Close()
	resp, err := conn.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

// Verify that HTTP (plaintext) round trip works via the simulation.
func TestDialHTTP_HTTP(t *testing.T) {
	nx := netcore.NewNetwork()
	req, err := http.NewRequestWithContext(t.Context(), "GET", "http://www.example.com/", nil)
	require.NoError(t, err)
	conn, err := nx.DialHTTP(req)
	require.NoError(t, err)
	defer conn.Close()
	resp, err := conn.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

// Verify that HTTPS round trip works when the URL contains an explicit port.
func TestDialHTTP_HTTPSExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	req, err := http.NewRequestWithContext(t.Context(), "GET", "https://www.example.com:443/", nil)
	require.NoError(t, err)
	conn, err := nx.DialHTTP(req)
	require.NoError(t, err)
	defer conn.Close()
	resp, err := conn.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

// Verify that HTTP round trip works when the URL contains an explicit port.
func TestDialHTTP_HTTPExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	req, err := http.NewRequestWithContext(t.Context(), "GET", "http://www.example.com:80/", nil)
	require.NoError(t, err)
	conn, err := nx.DialHTTP(req)
	require.NoError(t, err)
	defer conn.Close()
	resp, err := conn.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}
