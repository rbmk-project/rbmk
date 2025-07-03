// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/miekg/dns"
	"github.com/rbmk-project/rbmk/pkg/dns/dnscore"
	"github.com/rbmk-project/rbmk/pkg/x/netsim"
)

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over UDP.
func Example_dnsOverUDP() {
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

	// Create the client connection with the DNS server.
	conn, err := clientStack.DialContext(ctx, "udp", "8.8.8.8:53")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create the query to send
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = []dns.Question{{
		Name:   "dns.google.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}}

	// Perform the DNS round trip
	clientDNS := &dns.Client{}
	resp, _, err := clientDNS.ExchangeWithConnContext(ctx, query, &dns.Conn{Conn: conn})
	if err != nil {
		log.Fatal(err)
	}

	// Print the responses
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			fmt.Printf("%s\n", a.A.String())
		}
	}

	// Output:
	// 8.8.8.8
}

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over TCP.
func Example_dnsOverTCP() {
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

	// Create the client connection with the DNS server.
	conn, err := clientStack.DialContext(ctx, "tcp", "8.8.8.8:53")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create the query to send
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = []dns.Question{{
		Name:   "dns.google.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}}

	// Perform the DNS round trip
	clientDNS := &dns.Client{Net: "tcp"}
	resp, _, err := clientDNS.ExchangeWithConnContext(ctx, query, &dns.Conn{Conn: conn})
	if err != nil {
		log.Fatal(err)
	}

	// Print the responses
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			fmt.Printf("%s\n", a.A.String())
		}
	}

	// Output:
	// 8.8.8.8
}

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over TLS.
func Example_dnsOverTLS() {
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

	// Create the client connection with the DNS server.
	conn, err := clientStack.DialContext(ctx, "tcp", "8.8.8.8:853")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	tconn := tls.Client(conn, &tls.Config{
		RootCAs:    scenario.RootCAs(),
		NextProtos: []string{"dot"},
		ServerName: "dns.google",
	})
	defer tconn.Close()
	if err := tconn.HandshakeContext(ctx); err != nil {
		log.Fatal(err)
	}

	// Create the query to send
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = []dns.Question{{
		Name:   "dns.google.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}}

	// Perform the DNS round trip
	clientDNS := &dns.Client{Net: "tcp-tls"}
	resp, _, err := clientDNS.ExchangeWithConnContext(ctx, query, &dns.Conn{Conn: tconn})
	if err != nil {
		log.Fatal(err)
	}

	// Print the responses
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			fmt.Printf("%s\n", a.A.String())
		}
	}

	// Output:
	// 8.8.8.8
}

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over HTTPS.
func Example_dnsOverHTTPS() {
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

	// Create the dnscore transport and the server address
	txp := &dnscore.Transport{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: clientStack.DialContext,
				TLSClientConfig: &tls.Config{
					RootCAs:    scenario.RootCAs(),
					ServerName: "dns.google",
				},
			},
		},
	}
	serverAddr := dnscore.NewServerAddr(
		dnscore.ProtocolDoH, "https://8.8.8.8/dns-query")

	// Create the query to send
	query, err := dnscore.NewQuery("dns.google", dns.TypeA)
	if err != nil {
		log.Fatal(err)
	}

	// Perform the DNS round trip
	resp, err := txp.Query(ctx, serverAddr, query)
	if err != nil {
		log.Fatal(err)
	}

	// Print the responses
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			fmt.Printf("%s\n", a.A.String())
		}
	}

	// Output:
	// 8.8.8.8
}
