// SPDX-License-Identifier: GPL-3.0-or-later

package netcore_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/bassosimone/netstub"
	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/rbmk-project/rbmk/internal/qacore"
	"github.com/rbmk-project/rbmk/internal/testablenet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simulation is the qacore simulation used by all tests in this package.
var simulation = qacore.MustNewSimulation("testdata", qacore.ScenarioV4)

func TestMain(m *testing.M) {
	testablenet.DialContext.Set(simulation.DialContext)
	testablenet.LookupHost.Set(simulation.LookupHost)
	testablenet.RootCAs.Set(simulation.CertPool())
	os.Exit(m.Run())
}

// Verify that all fields are non-nil after construction.
func TestNewNetwork(t *testing.T) {
	nx := netcore.NewNetwork()
	assert.NotNil(t, nx.DialContextFunc)
	assert.NotNil(t, nx.Logger)
	assert.NotNil(t, nx.Resolver)
	assert.NotNil(t, nx.SplitHostPort)
	assert.NotNil(t, nx.TLSConfig)
	assert.NotNil(t, nx.TimeNow)
}

// Verify that dial returns error when SplitHostPort fails.
func TestDial_SplitHostPortError(t *testing.T) {
	errMocked := errors.New("mocked SplitHostPort error")
	nx := &netcore.Network{
		SplitHostPort: func(endpoint string) (string, string, error) {
			return "", "", errMocked
		},
	}
	_, err := nx.DialContext(t.Context(), "tcp", "www.example.com:443")
	require.ErrorIs(t, err, errMocked)
}

// Verify that dial returns error when LookupHost fails.
func TestDial_ResolverError(t *testing.T) {
	errMocked := errors.New("mocked resolver error")
	nx := &netcore.Network{
		SplitHostPort: func(endpoint string) (string, string, error) {
			return "www.example.com", "443", nil
		},
		Resolver: &netstub.FuncResolver{
			LookupHostFunc: func(ctx context.Context, name string) ([]string, error) {
				return nil, errMocked
			},
		},
	}
	_, err := nx.DialContext(t.Context(), "tcp", "www.example.com:443")
	require.ErrorIs(t, err, errMocked)
}

// Verify that dial returns error when the resolver returns unparseable addresses.
func TestDial_BadResolvedAddress(t *testing.T) {
	nx := netcore.NewNetwork()
	nx.SplitHostPort = func(endpoint string) (string, string, error) {
		return "www.example.com", "443", nil
	}
	nx.Resolver = &netstub.FuncResolver{
		LookupHostFunc: func(ctx context.Context, name string) ([]string, error) {
			return []string{"not-an-ip"}, nil
		},
	}
	_, err := nx.DialContext(t.Context(), "tcp", "www.example.com:443")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not-an-ip")
}

// Verify that TCP connect to www.example.com:443 succeeds.
func TestDialContext_TCP(t *testing.T) {
	nx := netcore.NewNetwork()
	conn, err := nx.DialContext(t.Context(), "tcp", "www.example.com:443")
	require.NoError(t, err)
	conn.Close()
}

// Verify that TCP connect to a port with no listener fails.
func TestDialContext_ConnectionRefused(t *testing.T) {
	nx := netcore.NewNetwork()
	_, err := nx.DialContext(t.Context(), "tcp", "www.example.com:9999")
	require.Error(t, err)
}

// Verify that TLS handshake to www.example.com:443 succeeds.
func TestDialTLSContext_Success(t *testing.T) {
	nx := netcore.NewNetwork()
	nx.TLSConfig.ServerName = "www.example.com"
	conn, err := nx.DialTLSContext(t.Context(), "tcp", "www.example.com:443")
	require.NoError(t, err)
	conn.Close()
}

// Verify that TLS verification fails with a wrong ServerName.
func TestDialTLSContext_WrongServerName(t *testing.T) {
	nx := netcore.NewNetwork()
	nx.TLSConfig.ServerName = "wrong.example.com"
	_, err := nx.DialTLSContext(t.Context(), "tcp", "www.example.com:443")
	require.Error(t, err)
}
