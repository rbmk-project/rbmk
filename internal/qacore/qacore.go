// SPDX-License-Identifier: GPL-3.0-or-later

// Package qacore contains the core QA functionality for RBMK.
package qacore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"sync"

	"github.com/bassosimone/dnstest"
	"github.com/bassosimone/minest"
	"github.com/bassosimone/pkitest"
	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/uis"
)

// DNSServer describes a DNS server in the [Scenario].
type DNSServer struct {
	// Addrs contains the IP addresses. Panics if empty.
	Addrs []netip.Addr

	// Domains contains the domain names to register in the DNS database.
	Domains []string

	// Aliases contains CNAME aliases pointing to Domains[0].
	// Panics if non-empty when Domains is empty.
	Aliases []string

	// Handler is the DNS handler. If nil, uses the global DNS database.
	Handler *dnstest.Handler
}

// ClientStack describes the client stack in the [Scenario].
type ClientStack struct {
	// Addrs contains the IP addresses. Panics if empty.
	Addrs []netip.Addr

	// Resolver is the DNS resolver address. Panics if zero value.
	Resolver netip.AddrPort
}

// HTTPServer describes an HTTP server in the [Scenario].
type HTTPServer struct {
	// Addrs contains the IP addresses. Panics if empty.
	Addrs []netip.Addr

	// Domains contains the domain names to register in the DNS database.
	Domains []string

	// Aliases contains CNAME aliases pointing to Domains[0].
	// Panics if non-empty when Domains is empty.
	Aliases []string

	// Handler is the HTTP handler. If nil, serves the default page.
	Handler http.Handler
}

// Scenario describes the simulated network topology.
type Scenario struct {
	// ClientStack describes the client stack.
	ClientStack ClientStack

	// DNSServers describes the DNS servers.
	DNSServers []DNSServer

	// HTTPServers describes the HTTP servers.
	HTTPServers []HTTPServer
}

// ScenarioV4 returns the IPv4 testing [*Scenario].
//
// This function returns a fresh [*Scenario] each time, so callers
// may mutate the result (e.g., set a custom Handler) before
// passing it to [MustNewSimulation].
func ScenarioV4() *Scenario {
	return &Scenario{
		ClientStack: ClientStack{
			Addrs:    []netip.Addr{netip.MustParseAddr("130.192.91.211")},
			Resolver: netip.MustParseAddrPort("130.192.3.21:53"),
		},
		DNSServers: []DNSServer{
			{
				Addrs:   []netip.Addr{netip.MustParseAddr("130.192.3.21")},
				Domains: []string{"giove.polito.it"},
			},
			{
				Addrs:   []netip.Addr{netip.MustParseAddr("8.8.4.4"), netip.MustParseAddr("8.8.8.8")},
				Domains: []string{"dns.google"},
				Aliases: []string{"dns.google.com"},
			},
		},
		HTTPServers: []HTTPServer{
			{
				Addrs:   []netip.Addr{netip.MustParseAddr("104.18.26.120")},
				Domains: []string{"www.example.com", "example.com"},
				Aliases: []string{"www.example.org", "example.org"},
			},
		},
	}
}

// PacketFilter filters packets.
type PacketFilter interface {
	// ShouldDrop may mutate the frame or drop it returning true.
	ShouldDrop(pkt uis.VNICFrame) bool
}

// PacketFilterFunc adapts a func to be a [PacketFilter].
type PacketFilterFunc func(pkt uis.VNICFrame) bool

var _ PacketFilter = PacketFilterFunc(nil)

// ShouldDrop implements [PacketFilter].
func (fx PacketFilterFunc) ShouldDrop(pkt uis.VNICFrame) bool {
	return fx(pkt)
}

// Router routes packets inside a [*Simulation].
type Router interface {
	// Route routes packets until ctx is canceled.
	Route(ctx context.Context, ix *uis.Internet)
}

// DefaultRouter is the default [Router] implementation.
//
// Use [NewDefaultRouter] to construct.
type DefaultRouter struct {
	// pf is the configured [PacketFilter].
	pf PacketFilter

	// pfmu protects the [PacketFilter].
	pfmu sync.RWMutex
}

// NewDefaultRouter creates a new [*DefaultRouter].
func NewDefaultRouter() *DefaultRouter {
	return &DefaultRouter{}
}

// SetPacketFilter sets a [PacketFilter] for [*DefaultRouter].
//
// Use nil to clear the [PacketFilter].
func (r *DefaultRouter) SetPacketFilter(pf PacketFilter) {
	r.pfmu.Lock()
	r.pf = pf
	r.pfmu.Unlock()
}

// Route implements [Router].
func (r *DefaultRouter) Route(ctx context.Context, ix *uis.Internet) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-ix.InFlight():
			r.pfmu.RLock()
			pf := r.pf
			r.pfmu.RUnlock()
			if pf == nil || !pf.ShouldDrop(frame) {
				_ = ix.Deliver(frame)
			}
		}
	}
}

// Simulation simulates the internet in QA tests.
//
// Use [MustNewSimulation] to construct. Cancel the context passed to
// [MustNewSimulation] to stop the simulation, then call [*Simulation.Wait]
// to wait for all goroutines to finish.
type Simulation struct {
	// dnsDB models the global, distributed DNS database.
	dnsDB *dnstest.HandlerConfig

	// pki contains the PKI used for testing.
	pki *pkitest.PKI

	// scenario contains the [*Scenario].
	scenario *Scenario

	// internet is the [*uis.Internet].
	internet *uis.Internet

	// userStack is the user [*uis.Stack].
	userStack *uis.Stack

	// closers collects things to close at shutdown time.
	closers []interface{ Close() }

	// wg tracks background goroutines.
	wg sync.WaitGroup
}

// MustNewSimulation creates a new [*Simulation].
//
// Cancel the ctx to stop the simulation, then call [*Simulation.Wait]
// to join background goroutines.
func MustNewSimulation(ctx context.Context, datadir string, scenario *Scenario, router Router) *Simulation {
	sx := &Simulation{
		dnsDB:    dnstest.NewHandlerConfig(),
		pki:      pkitest.MustNewPKI(datadir),
		scenario: scenario,
	}
	sx.mustInit(ctx, router)
	return sx
}

// Wait waits for all background goroutines to finish.
//
// You should cancel the context passed to [MustNewSimulation] before
// calling this method, otherwise it blocks forever.
func (sx *Simulation) Wait() {
	sx.wg.Wait()
}

// CertPool returns the cert pool that the user should use.
func (sx *Simulation) CertPool() *x509.CertPool {
	return sx.pki.CertPool()
}

// DialContext dials a connection using the [*uis.Stack] assigned to the user.
//
// This method is not able to resolve domain names and can only connect to TCP/UDP endpoints.
func (sx *Simulation) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return uis.NewConnector(sx.userStack).DialContext(ctx, network, address)
}

// LookupHost resolves a domain name using the [*uis.Stack] assigned to the user.
//
// This method uses the user resolver configured using a [ClientStack].
// Each call creates a fresh transport and resolver, which is acceptable for testing.
func (sx *Simulation) LookupHost(ctx context.Context, domain string) ([]string, error) {
	connector := uis.NewConnector(sx.userStack)
	txp := minest.NewDNSOverUDPTransport(connector, sx.scenario.ClientStack.Resolver)
	reso := minest.NewResolver(txp)
	return reso.LookupHost(ctx, domain)
}

// mustInit initializes the [*Simulation].
func (sx *Simulation) mustInit(ctx context.Context, router Router) {
	// Ensure we initialized dependencies correctly
	runtimex.Assert(sx.dnsDB != nil)
	runtimex.Assert(sx.pki != nil)
	runtimex.Assert(sx.scenario != nil)

	// Init the internet
	sx.internet = uis.NewInternet()

	// Init the client stack
	sx.mustInitClientStack()

	// Init each DNS server
	for idx := range sx.scenario.DNSServers {
		sx.mustInitDNSServer(&sx.scenario.DNSServers[idx])
	}

	// Init each HTTP server
	for idx := range sx.scenario.HTTPServers {
		sx.mustInitHTTPServer(&sx.scenario.HTTPServers[idx])
	}

	// Start the teardown goroutine
	sx.wg.Go(func() {
		router.Route(ctx, sx.internet)
		for idx := len(sx.closers) - 1; idx >= 0; idx-- {
			sx.closers[idx].Close()
		}
	})
}

// mustInitClientStack initializes the client stack.
func (sx *Simulation) mustInitClientStack() {
	cs := &sx.scenario.ClientStack
	runtimex.Assert(len(cs.Addrs) > 0)
	runtimex.Assert(cs.Resolver != netip.AddrPort{})
	sx.userStack = runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, cs.Addrs...,
	))
	sx.closers = append(sx.closers, sx.userStack)
}

// mustInitDNSServer initializes a [DNSServer].
func (sx *Simulation) mustInitDNSServer(ds *DNSServer) {
	// Sanity checks
	runtimex.Assert(len(ds.Addrs) > 0)
	if len(ds.Aliases) > 0 {
		runtimex.Assert(len(ds.Domains) > 0)
	}

	// Create the stack
	stack := runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, ds.Addrs...,
	))
	sx.closers = append(sx.closers, stack)

	// Register domains and aliases in the DNS database
	for _, domain := range ds.Domains {
		for _, ipAddr := range ds.Addrs {
			sx.dnsDB.AddNetipAddr(domain, ipAddr)
		}
	}
	for _, alias := range ds.Aliases {
		sx.dnsDB.AddCNAME(alias, ds.Domains[0])
	}

	// Determine the handler to use
	handler := ds.Handler
	if handler == nil {
		handler = dnstest.NewHandler(sx.dnsDB)
	}

	// Compute allDomains for the certificate (copy to avoid mutation)
	allDomains := make([]string, 0, len(ds.Domains)+len(ds.Aliases))
	allDomains = append(allDomains, ds.Domains...)
	allDomains = append(allDomains, ds.Aliases...)

	// Create a TLS certificate if we have domains
	var (
		cert    tls.Certificate
		hasCert bool
	)
	if len(ds.Domains) > 0 {
		ipAddrs := make([]net.IP, 0, len(ds.Addrs))
		for _, addr := range ds.Addrs {
			ipAddrs = append(ipAddrs, addr.AsSlice())
		}
		cert = sx.pki.MustNewCert(&pkitest.SelfSignedCertConfig{
			CommonName: ds.Domains[0],
			DNSNames:   allDomains,
			IPAddrs:    ipAddrs,
		})
		hasCert = true
	}

	// Create a listen config for this stack
	lcfg := uis.NewListenConfig(stack)

	// Start servers for each address
	for _, addr := range ds.Addrs {
		// UDP server
		udpSrv := dnstest.MustNewUDPServer(
			lcfg,
			makeStringEpnt(addr, 53),
			handler,
		)
		sx.closers = append(sx.closers, udpSrv)

		// TCP server
		tcpSrv := dnstest.MustNewTCPServer(
			lcfg,
			makeStringEpnt(addr, 53),
			handler,
		)
		sx.closers = append(sx.closers, tcpSrv)

		// TLS and HTTPS servers require a certificate
		if hasCert {
			tlsSrv := dnstest.MustNewTLSServer(
				lcfg,
				makeStringEpnt(addr, 853),
				cert,
				handler,
			)
			sx.closers = append(sx.closers, tlsSrv)

			httpsSrv := dnstest.MustNewHTTPSServer(
				lcfg,
				makeStringEpnt(addr, 443),
				cert,
				handler,
			)
			sx.closers = append(sx.closers, httpsSrv)
		}
	}
}

//go:embed example.com.html
var exampleComHTML string

// defaultWwwHandler is the default [http.Handler] for HTTP servers.
var defaultWwwHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(exampleComHTML))
})

// mustInitHTTPServer initializes an [HTTPServer].
func (sx *Simulation) mustInitHTTPServer(hs *HTTPServer) {
	// Sanity checks
	runtimex.Assert(len(hs.Addrs) > 0)
	if len(hs.Aliases) > 0 {
		runtimex.Assert(len(hs.Domains) > 0)
	}

	// Create the stack
	stack := runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, hs.Addrs...,
	))
	sx.closers = append(sx.closers, stack)

	// Register domains and aliases in the DNS database
	for _, domain := range hs.Domains {
		for _, ipAddr := range hs.Addrs {
			sx.dnsDB.AddNetipAddr(domain, ipAddr)
		}
	}
	for _, alias := range hs.Aliases {
		sx.dnsDB.AddCNAME(alias, hs.Domains[0])
	}

	// Determine the handler to use
	handler := hs.Handler
	if handler == nil {
		handler = defaultWwwHandler
	}

	// Compute allDomains for the certificate (copy to avoid mutation)
	allDomains := make([]string, 0, len(hs.Domains)+len(hs.Aliases))
	allDomains = append(allDomains, hs.Domains...)
	allDomains = append(allDomains, hs.Aliases...)

	// Create a TLS certificate if we have domains
	var cert tls.Certificate
	var hasCert bool
	if len(hs.Domains) > 0 {
		ipAddrs := make([]net.IP, 0, len(hs.Addrs))
		for _, addr := range hs.Addrs {
			ipAddrs = append(ipAddrs, addr.AsSlice())
		}
		cert = sx.pki.MustNewCert(&pkitest.SelfSignedCertConfig{
			CommonName: hs.Domains[0],
			DNSNames:   allDomains,
			IPAddrs:    ipAddrs,
		})
		hasCert = true
	}

	// Create a listen config for this stack
	lcfg := uis.NewListenConfig(stack)

	// Start servers for each address
	for _, addr := range hs.Addrs {
		// HTTP server on port 80
		httpListener := runtimex.PanicOnError1(lcfg.Listen(
			context.Background(),
			"tcp",
			makeStringEpnt(addr, 80),
		))
		httpSrv := httptest.NewUnstartedServer(handler)
		httpSrv.Listener = httpListener
		httpSrv.Start()
		sx.closers = append(sx.closers, httpSrv)

		// HTTPS server on port 443 requires a certificate
		if hasCert {
			httpsListener := runtimex.PanicOnError1(lcfg.Listen(
				context.Background(),
				"tcp",
				makeStringEpnt(addr, 443),
			))
			httpsSrv := httptest.NewUnstartedServer(handler)
			httpsSrv.Listener = httpsListener
			httpsSrv.TLS = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			httpsSrv.EnableHTTP2 = true
			httpsSrv.StartTLS()
			sx.closers = append(sx.closers, httpsSrv)
		}
	}
}

// makeStringEpnt constructs the string representation of an endpoint.
func makeStringEpnt(addr netip.Addr, port uint16) string {
	return netip.AddrPortFrom(addr, port).String()
}
