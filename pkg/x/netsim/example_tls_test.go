// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netsim"
)

// This example shows how to use [netsim] to simulate a TLS
// server that listens for incoming encrypted requests.
func Example_tls() {
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

	// Create and attach the client stack.
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create a context with a watchdog timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect to the server
	conn, err := clientStack.DialContext(ctx, "tcp", "8.8.8.8:443")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Perform the TLS handshake
	tconn := tls.Client(conn, &tls.Config{
		RootCAs:    scenario.RootCAs(),
		ServerName: "dns.google",
	})
	defer tconn.Close()
	if err := tconn.HandshakeContext(ctx); err != nil {
		log.Fatal(err)
	}

	// Print the handshake result
	fmt.Printf("%v", err)

	// Output:
	// <nil>
}
