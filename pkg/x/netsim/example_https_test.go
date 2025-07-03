// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rbmk-project/rbmk/pkg/x/netsim"
)

// This example shows how to use [netsim] to simulate an HTTPS
// server that listens for incoming encrypted requests.
func Example_https() {
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

	// Create the HTTP client
	clientTxp := scenario.NewHTTPTransport(clientStack)
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("https://8.8.8.8/")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Printf("%s", string(body))

	// Output:
	// Google Public DNS server.
}
