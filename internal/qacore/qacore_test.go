// SPDX-License-Identifier: GPL-3.0-or-later

package qacore_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"

	"github.com/rbmk-project/rbmk/internal/qacore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchWwwExampleCom fetches https://www.example.com/ through the simulation
// and returns the response status code and body.
func fetchWwwExampleCom(t *testing.T) (int, string) {
	t.Helper()
	ctx := t.Context()

	// Resolve www.example.com
	_, err := simulation.LookupHost(ctx, "www.example.com")
	require.NoError(t, err)

	// Get www.example.com
	txp := &http.Transport{
		DialContext:     simulation.DialContext,
		TLSClientConfig: &tls.Config{RootCAs: simulation.CertPool()},
	}
	clnt := &http.Client{Transport: txp}
	hr, err := clnt.Get("https://www.example.com/")
	require.NoError(t, err)
	defer hr.Body.Close()

	body, err := io.ReadAll(hr.Body)
	require.NoError(t, err)
	return hr.StatusCode, string(body)
}

// fetchWwwExampleComHTTP fetches http://www.example.com/ (plaintext)
// through the simulation and returns the response status code and body.
func fetchWwwExampleComHTTP(t *testing.T) (int, string) {
	t.Helper()
	ctx := t.Context()

	// Resolve www.example.com
	_, err := simulation.LookupHost(ctx, "www.example.com")
	require.NoError(t, err)

	// Get www.example.com
	txp := &http.Transport{
		DialContext: simulation.DialContext,
	}
	clnt := &http.Client{Transport: txp}
	hr, err := clnt.Get("http://www.example.com/")
	require.NoError(t, err)
	defer hr.Body.Close()

	body, err := io.ReadAll(hr.Body)
	require.NoError(t, err)
	return hr.StatusCode, string(body)
}

// Verify that the HTTP (plaintext) server on port 80 serves the same
// default content as the HTTPS server.
func TestHTTPWwwExampleCom(t *testing.T) {
	status, body := fetchWwwExampleComHTTP(t)
	assert.Equal(t, 200, status)
	assert.Equal(t, 605, len(body))
}

// Verify that a custom HTTP handler can be set at construction time
// via the Scenario.
func TestCustomHTTPHandler(t *testing.T) {
	const customBody = "custom response body"

	// Build a scenario with a custom handler for the HTTP server
	scenario := qacore.ScenarioV4()
	scenario.HTTPServers[0].Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(customBody))
	})

	// Create and tear down a dedicated simulation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := qacore.NewDefaultRouter()
	sim := qacore.MustNewSimulation(ctx, "testdata", scenario, r)
	defer func() {
		cancel()
		sim.Wait()
	}()

	// Resolve and fetch
	_, err := sim.LookupHost(t.Context(), "www.example.com")
	require.NoError(t, err)

	txp := &http.Transport{
		DialContext:     sim.DialContext,
		TLSClientConfig: &tls.Config{RootCAs: sim.CertPool()},
	}
	clnt := &http.Client{Transport: txp}
	hr, err := clnt.Get("https://www.example.com/")
	require.NoError(t, err)
	defer hr.Body.Close()

	body, err := io.ReadAll(hr.Body)
	require.NoError(t, err)
	assert.Equal(t, 200, hr.StatusCode)
	assert.Equal(t, customBody, string(body))
}
