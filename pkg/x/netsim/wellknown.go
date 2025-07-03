//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Well-known host configurations for common internet services
// used in network testing scenarios.
//

package netsim

import "net/http"

// MustNewGoogleDNSStack creates a new stack simulating dns.google.
func (s *Scenario) MustNewGoogleDNSStack() *Stack {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Google Public DNS server.\n"))
	}))
	mux.Handle("/dns-query", NewDNSHTTPHandler(*s.dnsd))
	return s.MustNewStack(&StackConfig{
		DomainNames: []string{
			"dns.google",
			"dns.google.com",
		},
		Addresses: []string{
			"2001:4860:4860::8888",
			"8.8.8.8",
		},
		DNSOverUDPHandler: s.DNSHandler(),
		DNSOverTCPHandler: s.DNSHandler(),
		DNSOverTLSHandler: s.DNSHandler(),
		HTTPSHandler:      mux,
	})
}

// MustNewExampleComStack creates a new stack simulating www.example.com.
func (s *Scenario) MustNewExampleComStack() *Stack {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Example Web Server.\n"))
	})
	return s.MustNewStack(&StackConfig{
		DomainNames: []string{
			"www.example.com",
			"example.com",
			"www.example.org",
			"example.org",
		},
		Addresses: []string{
			"2606:2800:21f:cb07:6820:80da:af6b:8b2c",
			"93.184.216.34",
		},
		HTTPHandler:  handler,
		HTTPSHandler: handler,
	})
}

// MustNewClientStack creates a new client stack with standard testing configuration.
//
// We use GARR's (Italian Research & Education Network) public addresses
// (193.206.158.22 and 2001:760:0:158::22) as default client addresses.
// These are chosen over documentation ranges (like 192.0.2.0/24) to avoid
// triggering bogon filters in network simulation scenarios, while still
// being associated with a public research institution.
//
// The stack uses Google's public DNS addresses as the default resolvers.
func (s *Scenario) MustNewClientStack() *Stack {
	return s.MustNewStack(&StackConfig{
		Addresses: []string{
			"193.206.158.22",
			"2001:760:0:158::22",
		},
		ClientResolvers: []string{
			"2001:4860:4860::8888",
			"8.8.8.8",
		},
	})
}

// MustNewBlockpageStack creates a new stack simulating a censorship blockpage server.
//
// It serves a simple warning page on HTTP/HTTPS indicating that the content has been blocked.
func (s *Scenario) MustNewBlockpageStack() *Stack {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access to this website has been blocked by network policy.\n"))
	})

	return s.MustNewStack(&StackConfig{
		Addresses: []string{
			"10.10.34.35",
		},
		HTTPHandler: handler,
	})
}
