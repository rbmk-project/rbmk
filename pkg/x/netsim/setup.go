// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/netip"

	"github.com/bassosimone/runtimex"
	"github.com/rbmk-project/rbmk/pkg/dns/dnscoretest"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/simpki"
)

// StackConfig contains configuration for creating a new network stack.
type StackConfig struct {
	// Addresses contains the IP addresses for this stack.
	//
	// The config is invalid if there is not at least one address.
	Addresses []string

	// ClientResolvers optionally specifies resolvers for client stacks.
	ClientResolvers []string

	// DNSOverUDPHandler optionally specifies a handler for DNS-over-UDP.
	DNSOverUDPHandler DNSHandler

	// DNSOverTCPHandler optionally specifies a handler for DNS-over-TCP.
	DNSOverTCPHandler DNSHandler

	// DNSOverTLSHandler optionally specifies a handler for DNS-over-TLS.
	DNSOverTLSHandler DNSHandler

	// DomainNames contains the optional domain names associated with this stack.
	//
	// If there are associated domain names, we will configure the DNS and
	// register related certificates for emulating the PKI.
	DomainNames []string

	// HTTPHandler optionally specifies a handle to use on port 80/tcp.
	HTTPHandler http.Handler

	// HTTPSHandler optionally specifies a handle to use on port 443/tcp.
	HTTPSHandler http.Handler
}

// validate returns an error if the configuration is not valid.
func (cfg *StackConfig) validate() error {
	if len(cfg.Addresses) < 1 {
		return errors.New("at least one address is required")
	}
	return nil
}

// newBaseStack returns the base stack given a [*StackConfig].
func (s *Scenario) newBaseStack(cfg *StackConfig) (*Stack, error) {
	addrs := make([]netip.Addr, len(cfg.Addresses))
	for idx, addr := range cfg.Addresses {
		pa, err := netip.ParseAddr(addr)
		if err != nil {
			return nil, err
		}
		addrs[idx] = pa
	}
	stack := NewStack(addrs...)
	return stack, nil
}

// setupClientResolvers configures the client resolvers for the stack.
func (cfg *StackConfig) setupClientResolvers(stack *Stack) error {
	var ress []netip.AddrPort
	for _, addr := range cfg.ClientResolvers {
		paddr, err := netip.ParseAddrPort(net.JoinHostPort(addr, "53"))
		if err != nil {
			return err
		}
		ress = append(ress, paddr)
	}
	stack.SetResolvers(ress...)
	return nil
}

// mustSetupPKI sets up the PKI database for the stack, if possible.
//
// This method panics on error.
func (s *Scenario) mustSetupPKI(cfg *StackConfig) (tls.Certificate, bool) {
	if len(cfg.DomainNames) <= 0 {
		return tls.Certificate{}, false
	}
	var ipAddr []net.IP
	for _, addr := range cfg.Addresses {
		ipAddr = append(ipAddr, netip.MustParseAddr(addr).AsSlice())
	}
	cert := s.pki.MustNewCert(&simpki.Config{
		CommonName: cfg.DomainNames[0],
		DNSNames:   cfg.DomainNames,
		IPAddrs:    ipAddr,
	})
	return cert, true
}

// mustSetupDNSOverUDP configures the DNS-over-UDP handler for the stack.
func (s *Scenario) mustSetupDNSOverUDP(stack *Stack, cfg *StackConfig) {
	server := &dnscoretest.Server{
		ListenPacket: func(network, address string) (net.PacketConn, error) {
			return stack.ListenPacket(context.Background(), network, "[::]:53")
		},
	}
	<-server.StartUDP(cfg.DNSOverUDPHandler)
	s.pool.Add(server)
}

// mustSetupDNSOverTCP configures the DNS-over-TCP handler for the stack.
func (s *Scenario) mustSetupDNSOverTCP(stack *Stack, cfg *StackConfig) {
	server := &dnscoretest.Server{
		Listen: func(network, address string) (net.Listener, error) {
			return stack.Listen(context.Background(), network, "[::]:53")
		},
	}
	<-server.StartTCP(cfg.DNSOverTCPHandler)
	s.pool.Add(server)
}

// mustSetupDNSOverTLS configures the DNS-over-TLS handler for the stack.
func (s *Scenario) mustSetupDNSOverTLS(stack *Stack, cfg *StackConfig, cert tls.Certificate) {
	server := &dnscoretest.Server{
		ListenTLS: func(network, address string, config *tls.Config) (net.Listener, error) {
			listener, err := stack.Listen(context.Background(), network, "[::]:853")
			if err != nil {
				return nil, err
			}
			config = config.Clone()
			config.Certificates = []tls.Certificate{cert}
			listener = tls.NewListener(listener, config)
			return listener, nil
		},
	}
	<-server.StartTLS(cfg.DNSOverTCPHandler)
	s.pool.Add(server)
}

// mustSetupHTTPOverTCP configures the HTTP-over-TCP handler for the stack.
func (s *Scenario) mustSetupHTTPOverTCP(stack *Stack, cfg *StackConfig) {
	listener := runtimex.PanicOnError1(stack.Listen(context.Background(), "tcp", "[::]:80"))
	srv := &http.Server{Handler: cfg.HTTPHandler}
	go srv.Serve(listener)
}

// mustSetupHTTPOverTLS configures the HTTP-over-TLS handler for the stack.
func (s *Scenario) mustSetupHTTPOverTLS(stack *Stack, cfg *StackConfig, cert tls.Certificate) {
	listener := runtimex.PanicOnError1(stack.Listen(context.Background(), "tcp", "[::]:443"))
	srv := &http.Server{
		Handler: cfg.HTTPSHandler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	go srv.ServeTLS(listener, "", "")
}
