// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netsim"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/geolink"
)

// This example shows how to use a router to simulate a network
// topology consisting of a client and multiple servers.
func Example_router() {
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

	// Create server stack emulating www.example.com.
	scenario.Attach(scenario.MustNewExampleComStack())

	// Create the client stack, build a geographic point-to-point link
	// and attach the scenario router to the other end of the link.
	clientStack := scenario.MustNewClientStack()
	linkDev := geolink.Extend(clientStack, &geolink.Config{
		Delay: 10 * time.Millisecond,
		Log:   true,
	})
	scenario.Attach(linkDev)

	// Create the HTTP client
	clientTxp := scenario.NewHTTPTransport(clientStack)
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("https://www.example.com/")
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
	// Example Web Server.
}
