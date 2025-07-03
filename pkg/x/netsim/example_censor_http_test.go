// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"

	"github.com/rbmk-project/rbmk/pkg/x/netsim"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/censor"
)

// This example shows how to use [netsim] to simulate transparent
// proxying of HTTP requests to serve blockpages.
func Example_blockpageTransparentProxy() {
	// Create scenario
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create blockpage server
	blockpage := scenario.MustNewBlockpageStack()
	scenario.Attach(blockpage)

	// Create target website
	scenario.Attach(scenario.MustNewExampleComStack())

	// Configure DNAT to send blocked traffic to blockpage server
	scenario.Router().AddFilter(censor.NewDNatter(
		netip.MustParseAddr("193.206.158.22"),       // source addr
		netip.MustParseAddrPort("93.184.216.34:80"), // target dest epnt
		netip.MustParseAddrPort("10.10.34.35:80"),   // repl dest epnt
	))

	// Create client stack
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create the HTTP client
	clientTxp := scenario.NewHTTPTransport(clientStack)
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("http://93.184.216.34/")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden {
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
	// Access to this website has been blocked by network policy.
}
