//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TLS config code.
//

package netcore

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

// tlsConfig either returns the (cloned) [*tls.Config] from the [Network] or
// creates a new one by invoking the [newTLSConfig] function.
func (nx *Network) tlsConfig(network, address string) (*tls.Config, error) {
	if nx.TLSConfig != nil {
		config := nx.TLSConfig.Clone() // make sure we return a cloned config
		return config, nil
	}
	return newTLSConfig(network, address, nx.RootCAs)
}

// newTLSConfig is a best-effort attempt at creating a suitable TLS config
// for TCP and UDP transports using the network and address.
func newTLSConfig(network, address string, pool *x509.CertPool) (*tls.Config, error) {
	sni, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		RootCAs:    pool, // default to nil, which implies using the system root CAs
		NextProtos: []string{},
		ServerName: sni,
	}
	switch {
	case port == "443" && network == "tcp":
		config.NextProtos = []string{"h2", "http/1.1"}
	case port == "443" && network == "udp":
		config.NextProtos = []string{"h3"}
	case port == "853" && network == "tcp":
		config.NextProtos = []string{"doh"}
	case port == "853" && network == "udp":
		config.NextProtos = []string{"doq"}
	}

	return config, nil
}
