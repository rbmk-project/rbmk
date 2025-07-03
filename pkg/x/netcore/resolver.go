//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
//
// Internal code for DNS lookups.
//

package netcore

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/rbmk-project/common/errclass"
)

// maybeLookupEndpoint resolves the domain name inside an endpoint into
// a list of TCP/UDP endpoints. If the domain name is already an IP
// address, we short circuit the lookup.
func (nx *Network) maybeLookupEndpoint(ctx context.Context, endpoint string) ([]string, error) {
	domain, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}

	addrs, err := nx.maybeLookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}

	var endpoints []string
	for _, addr := range addrs {
		endpoints = append(endpoints, net.JoinHostPort(addr, port))
	}
	return endpoints, nil
}

// maybeLookupHost resolves a domain name to IP addresses unless the domain
// is already an IP address, in which case we short circuit the lookup.
func (nx *Network) maybeLookupHost(ctx context.Context, domain string) ([]string, error) {
	// handle the case where domain is already an IP address
	if net.ParseIP(domain) != nil {
		return []string{domain}, nil
	}

	// Optionally enforce a timeout for the lookup
	if nx.LookupHostTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, nx.LookupHostTimeout)
		defer cancel()
	}

	// Emit structured event before the lookup
	t0 := nx.emitLookupHostStart(ctx, domain)

	// Perform the actual lookup
	addrs, err := nx.doLookupHost(ctx, domain)

	// Emit structured event after the lookup
	nx.emitLookupHostDone(ctx, domain, t0, addrs, err)

	// Returns results to the caller
	return addrs, err
}

// defaultResolver is the [*net.Resolver] we use by default.
var defaultResolver = &net.Resolver{}

// doLookupHost performs the DNS lookup.
func (nx *Network) doLookupHost(ctx context.Context, domain string) ([]string, error) {
	// if there is a custom LookupHostFunc, use it
	if nx.LookupHostFunc != nil {
		return nx.LookupHostFunc(ctx, domain)
	}

	// otherwise either use the default [*net.Resolver] or the
	// default override through NewResolverOrSingleton
	reso := defaultResolver
	if nx.NewResolverOrSingleton != nil {
		reso = nx.NewResolverOrSingleton()
	}
	return reso.LookupHost(ctx, domain)
}

// emitLookupHostStart emits a structured event before the lookup.
func (nx *Network) emitLookupHostStart(ctx context.Context, domain string) time.Time {
	t0 := nx.timeNow()
	if nx.Logger != nil {
		nx.Logger.InfoContext(
			ctx,
			"lookupHostStart",
			slog.String("dnsLookupDomain", domain),
			slog.Time("t", t0),
		)
	}
	return t0
}

// emitLookupHostDone emits a structured event after the lookup.
func (nx *Network) emitLookupHostDone(ctx context.Context,
	domain string, t0 time.Time, addrs []string, err error) {
	if nx.Logger != nil {
		nx.Logger.InfoContext(
			ctx,
			"lookupHostDone",
			slog.String("dnsLookupDomain", domain),
			slog.Any("dnsResolvedAddrs", addrs),
			slog.Any("err", err),
			slog.String("errClass", errclass.New(err)),
			slog.Time("t0", t0),
			slog.Time("t", nx.timeNow()),
		)
	}
}
