// SPDX-License-Identifier: GPL-3.0-or-later

package netcore_test

import (
	"net/url"
	"testing"

	"github.com/bassosimone/dnscodec"
	"github.com/miekg/dns"
	"github.com/rbmk-project/rbmk/internal/netcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dnsExchangeA performs a DNS A query for the given domain over the
// given [netcore.DNSConn] and returns the resolved addresses.
func dnsExchangeA(t *testing.T, conn netcore.DNSConn, domain string) []string {
	t.Helper()
	query := dnscodec.NewQuery(domain, dns.TypeA)
	resp, err := conn.Exchange(t.Context(), query)
	require.NoError(t, err)
	addrs, err := resp.RecordsA()
	require.NoError(t, err)
	return addrs
}

// Verify that DialDNS returns error for an unsupported scheme.
func TestDialDNS_UnsupportedScheme(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("ftp://dns.google")
	require.NoError(t, err)
	_, err = nx.DialDNS(t.Context(), u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme")
}

// Verify that DNS-over-UDP works via the simulation.
func TestDialDNS_UDP(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("udp://dns.google")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-TCP works via the simulation.
func TestDialDNS_TCP(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("tcp://dns.google")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-TLS works via the simulation.
func TestDialDNS_DoT(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("dot://dns.google")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-HTTPS works via the simulation.
func TestDialDNS_DoH(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("https://dns.google/dns-query")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-UDP works when the URL contains an explicit port.
func TestDialDNS_UDPExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("udp://dns.google:53")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-TCP works when the URL contains an explicit port.
func TestDialDNS_TCPExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("tcp://dns.google:53")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-TLS works when the URL contains an explicit port.
func TestDialDNS_DoTExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("dot://dns.google:853")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}

// Verify that DNS-over-HTTPS works when the URL contains an explicit port.
func TestDialDNS_DoHExplicitPort(t *testing.T) {
	nx := netcore.NewNetwork()
	u, err := url.Parse("https://dns.google:443/dns-query")
	require.NoError(t, err)
	conn, err := nx.DialDNS(t.Context(), u)
	require.NoError(t, err)
	defer conn.Close()
	addrs := dnsExchangeA(t, conn, "www.example.com")
	assert.Contains(t, addrs, "104.18.26.120")
}
