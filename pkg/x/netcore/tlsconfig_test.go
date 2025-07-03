// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetwork_tlsConfig(t *testing.T) {
	t.Run("returns cloned config when available", func(t *testing.T) {
		originalConfig := &tls.Config{
			ServerName:         "example.com",
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}

		nx := &Network{
			TLSConfig: originalConfig,
		}

		config, err := nx.tlsConfig("tcp", "example.com:443")
		require.NoError(t, err)

		// Verify it's not the same instance (it's cloned)
		assert.NotSame(t, originalConfig, config)

		// But has the same values
		assert.Equal(t, originalConfig.ServerName, config.ServerName)
		assert.Equal(t, originalConfig.InsecureSkipVerify, config.InsecureSkipVerify)
		assert.Equal(t, originalConfig.MinVersion, config.MinVersion)
	})

	t.Run("creates new config when none is available", func(t *testing.T) {
		nx := &Network{}

		config, err := nx.tlsConfig("tcp", "example.com:443")
		require.NoError(t, err)

		// Verify basic properties of the created config
		assert.Equal(t, "example.com", config.ServerName)
		assert.False(t, config.InsecureSkipVerify)
		assert.Contains(t, config.NextProtos, "h2")
		assert.Contains(t, config.NextProtos, "http/1.1")
	})

	t.Run("passes root CAs to newTLSConfig", func(t *testing.T) {
		// Create a mock cert pool
		pool := x509.NewCertPool()

		nx := &Network{
			RootCAs: pool,
		}

		config, err := nx.tlsConfig("tcp", "example.com:443")
		require.NoError(t, err)

		// Verify the root CAs were passed through
		assert.Same(t, pool, config.RootCAs)
	})
}

func TestNewTLSConfig(t *testing.T) {
	t.Run("invalid address format", func(t *testing.T) {
		_, err := newTLSConfig("tcp", "invalid-address", nil)
		assert.Error(t, err)
	})

	t.Run("basic tcp:443 config", func(t *testing.T) {
		config, err := newTLSConfig("tcp", "example.com:443", nil)
		require.NoError(t, err)

		assert.Equal(t, "example.com", config.ServerName)
		assert.ElementsMatch(t, []string{"h2", "http/1.1"}, config.NextProtos)
	})

	t.Run("udp:443 for QUIC/HTTP3", func(t *testing.T) {
		config, err := newTLSConfig("udp", "example.com:443", nil)
		require.NoError(t, err)

		assert.Equal(t, "example.com", config.ServerName)
		assert.ElementsMatch(t, []string{"h3"}, config.NextProtos)
	})

	t.Run("tcp:853 for DoT (DNS over TLS)", func(t *testing.T) {
		config, err := newTLSConfig("tcp", "dns.example.com:853", nil)
		require.NoError(t, err)

		assert.Equal(t, "dns.example.com", config.ServerName)
		assert.ElementsMatch(t, []string{"doh"}, config.NextProtos)
	})

	t.Run("tcp:853 for DoT (DNS over TLS)", func(t *testing.T) {
		config, err := newTLSConfig("udp", "dns.example.com:853", nil)
		require.NoError(t, err)

		assert.Equal(t, "dns.example.com", config.ServerName)
		assert.ElementsMatch(t, []string{"doq"}, config.NextProtos)
	})

	t.Run("custom port with no special ALPN", func(t *testing.T) {
		config, err := newTLSConfig("tcp", "example.com:8443", nil)
		require.NoError(t, err)

		assert.Equal(t, "example.com", config.ServerName)
		assert.Empty(t, config.NextProtos)
	})

	t.Run("passes custom root CAs", func(t *testing.T) {
		pool := x509.NewCertPool()

		config, err := newTLSConfig("tcp", "example.com:443", pool)
		require.NoError(t, err)

		assert.Same(t, pool, config.RootCAs)
	})
}
