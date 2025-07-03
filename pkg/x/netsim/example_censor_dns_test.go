// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
	"github.com/rbmk-project/rbmk/pkg/dns/dnscore"
	"github.com/rbmk-project/rbmk/pkg/x/netsim"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/censor"
	netsimdns "github.com/rbmk-project/rbmk/pkg/x/netsim/dns"
)

// This example shows how to use [netsim] to simulate GFW-style DNS
// censorship, where poisoned responses are injected before the legitimate
// response arrives. The example demonstrates:
//
// 1. how to configure DNS poisoning using a database
// 2. how to collect multiple DNS responses using dnscore
// 3. the expected order of responses (poisoned then legitimate)
//
// This example DOES NOT show how to validate responses using [dnscore]
// since that is outside its specific scope.
func Example_censorDNS() {
	// Create a new scenario using the given directory to cache
	// the certificates used by the simulated PKI
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create server stack emulating dns.google (8.8.8.8).
	//
	// This includes:
	//
	// 1. creating, attaching, and enabling routing for a server stack
	//
	// 2. registering the proper domain names and addresses
	//
	// 3. updating the PKI database to include the server's certificate
	scenario.Attach(scenario.MustNewGoogleDNSStack())

	// Configure DNS poisoning happening on the scenario router
	// thus modeling the typical behaviour of the GFW.
	censorDB := netsimdns.NewDatabase()
	censorDB.AddAddresses([]string{"dns.google"}, []string{"10.0.0.1"})
	scenario.Router().AddFilter(censor.NewDNSPoisoner(censorDB))

	// Create and attach the client stack.
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create a context with a watchdog timeout.
	//
	// In real measurements this would typically be controlled by
	// --wait-duplicates or natural timing of other operations like
	// TCP/TLS handshakes and fetching related web pages.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create DNS query for dns.google A record
	query, err := dnscore.NewQuery("dns.google.", dns.TypeA)
	if err != nil {
		log.Fatal(err)
	}

	// Configure transport to use our simulated network
	txp := &dnscore.Transport{}
	txp.DialContext = clientStack.DialContext

	// Query 8.8.8.8 over UDP and collect responses
	serverAddr := &dnscore.ServerAddr{
		Protocol: dnscore.ProtocolUDP,
		Address:  "8.8.8.8:53",
	}
	results := txp.QueryWithDuplicates(ctx, serverAddr, query)

	// Print responses as they arrive.
	//
	// We expect:
	//
	// 1. poisoned response (10.0.0.1) from router
	//
	// 2. legitimate response (8.8.8.8) from server
	//
	// After two responses, we cancel the context. In production,
	// we will stop after a timeout or perform other operations and
	// then check whether there are more addresses to measure.
	var count int
	for result := range results {
		if err := result.Err; err != nil {
			// Errors here typically are caused by us closing
			// the connection and, anyway, for this test we only
			// care about seeing the duplicate responses.
			break
		}
		for _, ans := range result.Msg.Answer {
			if a, ok := ans.(*dns.A); ok {
				fmt.Printf("%s\n", a.A.String())
			}
		}
		count++
		if count >= 2 {
			cancel()
		}
	}

	// Output:
	// 10.0.0.1
	// 8.8.8.8
}
