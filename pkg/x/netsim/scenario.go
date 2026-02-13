// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"crypto/x509"

	"github.com/bassosimone/closepool"
	"github.com/bassosimone/pkitest"
	"github.com/bassosimone/runtimex"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/packet"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/router"
)

// Scenario manages network simulation components using a star topology,
// where all stacks are connected through a central router.
//
// This means:
//
// 1. Each stack is connected only to the central router;
//
// 2. The router forwards packets between stacks.
type Scenario struct {
	// dnsd is the [*DNSDatabase].
	dnsd *dnsDatabase

	// pki is the [*PKI] database.
	pki *pkitest.PKI

	// pool tracks all that which needs to be closed.
	pool *closepool.Pool

	// router is the star-topology router.
	router *router.Router
}

// NewScenario creates a new network simulation scenario.
//
// The cacheDir caches simulated-PKI-related data.
func NewScenario(cacheDir string) *Scenario {
	return &Scenario{
		dnsd:   newDNSDatabase(),
		pki:    pkitest.MustNewPKI(cacheDir),
		pool:   &closepool.Pool{},
		router: router.New(),
	}
}

// Router returns the [*router.Router] for the scenario.
func (s *Scenario) Router() *router.Router {
	return s.router
}

// DNSHandler returns the [DNSHandler] for the scenario. The returned
// handler will serve queries based on the scenario's DNS database.
func (s *Scenario) DNSHandler() DNSHandler {
	return s.dnsd
}

// RootCAs returns the [*x509.CertPool] that clients should use.
func (s *Scenario) RootCAs() *x509.CertPool {
	return s.pki.CertPool()
}

// MustNewStack creates a new network stack using the given configuration.
//
// This method panics on error.
//
// This method IS NOT goroutine safe.
func (s *Scenario) MustNewStack(config *StackConfig) *Stack {
	// Initialize and configure the stack.
	runtimex.PanicOnError0(config.validate())
	stack := runtimex.PanicOnError1(s.newBaseStack(config))
	runtimex.PanicOnError0(config.setupClientResolvers(stack))
	s.dnsd.AddAddresses(config.DomainNames, config.Addresses)
	cert, hasCert := s.mustSetupPKI(config)

	// Start DNS handlers.
	if config.DNSOverUDPHandler != nil {
		s.mustSetupDNSOverUDP(stack, config)
	}
	if config.DNSOverTCPHandler != nil {
		s.mustSetupDNSOverTCP(stack, config)
	}
	if config.DNSOverTLSHandler != nil {
		runtimex.Assert(hasCert)
		s.mustSetupDNSOverTLS(stack, config, cert)
	}

	// Start HTTP handlers.
	if config.HTTPHandler != nil {
		s.mustSetupHTTPOverTCP(stack, config)
	}
	if config.HTTPSHandler != nil {
		runtimex.Assert(hasCert)
		s.mustSetupHTTPOverTLS(stack, config, cert)
	}

	s.pool.Add(stack)
	return stack
}

// Close releases all resources associated with the scenario.
func (s *Scenario) Close() error {
	return s.pool.Close()
}

// Attach connects a device to the scenario's central router to
// read packets from the device, and sets up the route to that
// return packets correctly reach the device.
//
// The common case is to attach a [*Stack] but other cases are also
// possible. Suppose a [*Stack] is linked to a firewall through a link,
// then you can also attach the firewall to the router.
//
// All network traffic to/from this device will flow through the router.
func (s *Scenario) Attach(dev packet.NetworkDevice) {
	s.router.Attach(dev)
}
