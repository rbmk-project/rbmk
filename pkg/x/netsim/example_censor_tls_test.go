// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netsim"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/censor"
)

// This example shows how to use [netsim] to simulate
// SNI-based TLS blocking using RST injection.
func Example_tlsRSTInjection() {
	// Create a new scenario using the given directory to cache
	// the certificates used by the simulated PKI
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create server stack emulating dns.google.
	//
	// This includes:
	//
	// 1. creating, attaching, and enabling routing for a server stack
	//
	// 2. registering the proper domain names and addresses
	//
	// 3. updating the PKI database to include the server's certificate
	scenario.Attach(scenario.MustNewGoogleDNSStack())

	// Configure RST injection on the scenario router targeting
	// connections where the SNI matches "dns.google"
	scenario.Router().AddFilter(censor.NewTCPResetter(
		netip.AddrPort{},     // match any endpoint
		[]byte("dns.google"), // match SNI
	))

	// Create and attach the client stack.
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create the HTTP client
	clientTxp := scenario.NewHTTPTransport(clientStack)
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{Transport: clientTxp}

	// Attempt the HTTPS request, which should fail due to RST
	_, err := clientHTTP.Get("https://dns.google/")
	fmt.Printf("err: %v\n", err)

	// Output:
	// err: Get "https://dns.google/": connection reset by peer
}

// This example shows how to use [netsim] to simulate SNI-based TLS blocking
// using connection blackholing.
func Example_tlsBlackholing() {
	// Create a new scenario using the given directory to cache
	// the certificates used by the simulated PKI
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create server stack emulating dns.google.
	scenario.Attach(scenario.MustNewGoogleDNSStack())

	// Configure blackholing on the scenario router targeting
	// connections where the SNI matches "dns.google"
	scenario.Router().AddFilter(censor.NewBlackholer(
		300*time.Second,      // residual censorship duration
		netip.AddrPort{},     // match any endpoint
		[]byte("dns.google"), // match SNI
	))

	// Create and attach the client stack.
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create the HTTP client with a short timeout
	clientTxp := scenario.NewHTTPTransport(clientStack)
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{
		Transport: clientTxp,
		Timeout:   200 * time.Millisecond, // short timeout for testing
	}

	// Attempt the HTTPS request, which should time out
	_, err := clientHTTP.Get("https://dns.google/")
	fmt.Printf("err: %v\n", err)

	// Output:
	// err: Get "https://dns.google/": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
}
