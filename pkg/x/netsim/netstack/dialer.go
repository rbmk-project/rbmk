//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Dialer implementation
//

package netstack

import (
	"context"
	"errors"
	"net"

	"github.com/rbmk-project/rbmk/pkg/dns/dnscore"
	"github.com/rbmk-project/x/netcore"
)

// ErrNoConfiguredResolvers is returned when there are no configured resolvers.
var ErrNoConfiguredResolvers = errors.New("no configured resolvers")

// DialContext dials a network address.
func (ns *Stack) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Short circuit for the case where we're dialing for an IP address
	ipAddr, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if net.ParseIP(ipAddr) != nil {
		return ns.dialContext(ctx, network, address)
	}

	// Otherwise, bail if there are no configured resolvers.
	if len(ns.resolvers) <= 0 {
		return nil, ErrNoConfiguredResolvers
	}

	// Configure dnscore and netcore to perform the actual dial.
	netx := &netcore.Network{}
	netx.DialContextFunc = ns.dialContext
	reso := &dnscore.Resolver{}
	reso.Config = dnscore.NewConfig()
	for _, server := range ns.resolvers {
		reso.Config.AddServer(server)
	}
	reso.Transport = &dnscore.Transport{
		DialContext: ns.dialContext,
	}
	netx.LookupHostFunc = reso.LookupHost
	return netx.DialContext(ctx, network, address)
}
