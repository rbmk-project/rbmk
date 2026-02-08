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

// AddressAssignments contains the IP address assignments.
type AddressAssignments struct {
	// DnsGoogle is the addresses assigned to dns.google.
	DnsGoogle netip.Addr

	// User is the address assigned to the user.
	User netip.Addr

	// UserResolver is the DNS-over-UDP resolver serving the user.
	UserResolver netip.AddrPort

	// WwwExampleCom is the address assigned to www.example.com.
	WwwExampleCom netip.Addr
}

// ScenarioV4 is the IPv4 testing scenario.
var ScenarioV4 = &AddressAssignments{
	DnsGoogle:     netip.MustParseAddr("8.8.4.4"),
	User:          netip.MustParseAddr("130.192.91.211"),
	UserResolver:  netip.MustParseAddrPort("130.192.3.21:53"),
	WwwExampleCom: netip.MustParseAddr("104.18.26.120"),
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

// Simulation simulates the internet in QA tests.
//
// Use [MustNewSimulation] to construct. A [*Simulation] is intended
// to be created once (typically as a package-level variable) and
// to live for the entire test process. Use [*Simulation.SetPacketFilter]
// to vary censorship conditions between subtests.
type Simulation struct {
	// dnsDB models the global, distributed DNS database.
	dnsDB *dnstest.HandlerConfig

	// dnsGoogleStack is the dns.google [*uis.Stack].
	dnsGoogleStack *uis.Stack

	// internet is the [*uis.Internet].
	internet *uis.Internet

	// pf is the configured [PacketFilter].
	pf PacketFilter

	// pfmu protects the [PacketFilter].
	pfmu sync.RWMutex

	// wwwHandler is the configurable [http.Handler] for www.example.com.
	wwwHandler http.Handler

	// wwwmu protects wwwHandler.
	wwwmu sync.RWMutex

	// pki contains the PKI used for testing.
	pki *pkitest.PKI

	// scenario contains the [AddressAssignments].
	scenario *AddressAssignments

	// userResolverStack is the user-resolver [*uis.Stack].
	userResolverStack *uis.Stack

	// userStack is the user [*uis.Stack].
	userStack *uis.Stack

	// wwwExampleComStack is the www.example.com [*uis.Stack].
	wwwExampleComStack *uis.Stack
}

// MustNewSimulation creates a new [*Simulation].
func MustNewSimulation(datadir string, scenario *AddressAssignments) *Simulation {
	sx := &Simulation{
		dnsDB:    dnstest.NewHandlerConfig(),
		pki:      pkitest.MustNewPKI(datadir),
		scenario: scenario,
	}
	sx.mustInit()
	return sx
}

// SetPacketFilter sets a [PacketFilter]. Use nil to clear the [PacketFilter].
func (sx *Simulation) SetPacketFilter(pf PacketFilter) {
	sx.pfmu.Lock()
	sx.pf = pf
	sx.pfmu.Unlock()
}

// SetWwwExampleComHandler sets the [http.Handler] for www.example.com.
// Use nil to revert to the default handler serving the embedded HTML page.
func (sx *Simulation) SetWwwExampleComHandler(handler http.Handler) {
	sx.wwwmu.Lock()
	sx.wwwHandler = handler
	sx.wwwmu.Unlock()
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
// This method uses the user resolver configured using an [*AddressAssignment].
// Each call creates a fresh transport and resolver, which is acceptable for testing.
func (sx *Simulation) LookupHost(ctx context.Context, domain string) ([]string, error) {
	connector := uis.NewConnector(sx.userStack)
	txp := minest.NewDNSOverUDPTransport(connector, sx.scenario.UserResolver)
	reso := minest.NewResolver(txp)
	return reso.LookupHost(ctx, domain)
}

// mustInit initializes the [*Simulation].
func (sx *Simulation) mustInit() {
	// Init the internet
	sx.internet = uis.NewInternet()

	// Ensure we initialized dependencies correctly
	runtimex.Assert(sx.dnsDB != nil)
	runtimex.Assert(sx.pki != nil)
	runtimex.Assert(sx.scenario != nil)

	// Init the user stack
	sx.userStack = runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, sx.scenario.User,
	))
	sx.dnsDB.AddNetipAddr("whitespider.polito.it", sx.scenario.User)

	// Init the user resolver stack
	sx.userResolverStack = runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, sx.scenario.UserResolver.Addr(),
	))
	sx.dnsDB.AddNetipAddr("giove.polito.it", sx.scenario.UserResolver.Addr())
	go sx.userResolverMain()

	// Init the dns.google stack
	sx.dnsGoogleStack = runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, sx.scenario.DnsGoogle,
	))
	sx.dnsDB.AddNetipAddr("dns.google", sx.scenario.DnsGoogle)
	dnsGoogleCertificate := sx.pki.MustNewCert(&pkitest.SelfSignedCertConfig{
		CommonName:   "dns.google",
		DNSNames:     []string{"dns.google"},
		IPAddrs:      []net.IP{sx.scenario.DnsGoogle.AsSlice()},
		Organization: []string{"Google"},
	})
	go sx.dnsGoogleMain(dnsGoogleCertificate)

	// Init the www.example.com stack
	sx.wwwExampleComStack = runtimex.PanicOnError1(sx.internet.NewStack(
		uis.MTUEthernet, sx.scenario.WwwExampleCom,
	))
	sx.dnsDB.AddNetipAddr("www.example.com", sx.scenario.WwwExampleCom)
	wwwExampleComCertificate := sx.pki.MustNewCert(&pkitest.SelfSignedCertConfig{
		CommonName:   "www.example.com",
		DNSNames:     []string{"www.example.com"},
		IPAddrs:      []net.IP{sx.scenario.WwwExampleCom.AsSlice()},
		Organization: []string{"Example"},
	})
	go sx.wwwExampleComMain(wwwExampleComCertificate)

	// Route packets
	go sx.route()
}

// userResolverMain is the main goroutine of the user resolver stack.
func (sx *Simulation) userResolverMain() {
	lcfg := uis.NewListenConfig(sx.userResolverStack)
	dnstest.MustNewUDPServer(
		lcfg,
		sx.scenario.UserResolver.String(),
		dnstest.NewHandler(sx.dnsDB),
	)
}

// dnsGoogleMain is the main goroutine of the dns.google stack.
func (sx *Simulation) dnsGoogleMain(certificate tls.Certificate) {
	lcfg := uis.NewListenConfig(sx.dnsGoogleStack)

	// UDP
	dnstest.MustNewUDPServer(
		lcfg,
		makeStringEpnt(sx.scenario.DnsGoogle, 53),
		dnstest.NewHandler(sx.dnsDB),
	)

	// TCP
	dnstest.MustNewTCPServer(
		lcfg,
		makeStringEpnt(sx.scenario.DnsGoogle, 53),
		dnstest.NewHandler(sx.dnsDB),
	)

	// TLS
	dnstest.MustNewTLSServer(
		lcfg,
		makeStringEpnt(sx.scenario.DnsGoogle, 853),
		certificate,
		dnstest.NewHandler(sx.dnsDB),
	)

	// HTTPS
	dnstest.MustNewHTTPSServer(
		lcfg,
		makeStringEpnt(sx.scenario.DnsGoogle, 443),
		certificate,
		dnstest.NewHandler(sx.dnsDB),
	)
}

//go:embed example.com.html
var exampleComHTML string

// defaultWwwHandler is the default [http.Handler] for www.example.com.
var defaultWwwHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(exampleComHTML))
})

// wwwExampleComMain is the main goroutine of the www.example.com stack.
func (sx *Simulation) wwwExampleComMain(certificate tls.Certificate) {
	// Listener
	lcfg := uis.NewListenConfig(sx.wwwExampleComStack)
	listener := runtimex.PanicOnError1(lcfg.Listen(
		context.Background(),
		"tcp",
		makeStringEpnt(sx.scenario.WwwExampleCom, 443),
	))

	// Handler that delegates to the configurable wwwHandler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sx.wwwmu.RLock()
		h := sx.wwwHandler
		sx.wwwmu.RUnlock()
		if h == nil {
			h = defaultWwwHandler
		}
		h.ServeHTTP(w, r)
	})

	// Server
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = listener
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	srv.EnableHTTP2 = true
	srv.StartTLS()
}

// route routes packets between hosts in the [*Simulation].
func (sx *Simulation) route() {
	for frame := range sx.internet.InFlight() {
		sx.pfmu.RLock()
		pf := sx.pf
		sx.pfmu.RUnlock()
		if pf == nil || !pf.ShouldDrop(frame) {
			_ = sx.internet.Deliver(frame)
		}
	}
}

// makeStringEpnt constructs the string representation of an endpoint.
func makeStringEpnt(addr netip.Addr, port uint16) string {
	return netip.AddrPortFrom(addr, port).String()
}
