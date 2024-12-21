//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Code to discover test helpers etc.
//

package generate

import (
	"context"
	"fmt"
	"net/netip"
	"regexp"

	"github.com/rbmk-project/dnscore"
)

// discoverer discovers test helpers addresses.
//
// The zero value IS NOT ready to use and you must explicitly call
// [newDiscoverer] to obtain a ready-to-use discoverer.
type discoverer struct {
	// disableIPv4 is true if we should disable IPv4.
	disableIPv4 bool

	// disableIPv6 is true if we should disable IPv6.
	disableIPv6 bool

	// enableSTUN is true if we should enable STUN.
	enableSTUN bool
}

// newDiscoverer creates a new [*discoverer] with the given settings.
func newDiscoverer(disableIPv4, disableIPv6, enableSTUN bool) *discoverer {
	return &discoverer{
		disableIPv4: disableIPv4,
		disableIPv6: disableIPv6,
		enableSTUN:  enableSTUN,
	}
}

// STUNAddrs discovers the STUN addresses to use. This function
// is like TestHelperAddrs but for STUN servers and does not run
// when STUN is disabled (i.e., enableSTUN is set to false).
func (d *discoverer) STUNAddrs(
	ctx context.Context, stunDomain string) (v4 string, v6 string, err error) {
	// 1. if STUN is disabled, return immediately
	if !d.enableSTUN {
		return "", "", nil
	}

	// 2. otherwise, fallback to the TestHelperAddrs function
	return d.TestHelperAddrs(ctx, stunDomain)
}

// domainRegexp is a regular expression to match domain names.
var domainRegexp = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+$`)

/*
TestHelperAddrs discovers the test helper addresses to use.

Arguments:

1. ctx is the context.

2. thDomain is the domain name of the test helper.

Notes:

1. We return the first IPv4 and IPv6 addresses found.

2. We fail if we do not find any suitable addresses (which may depend
either on DNS results or on the disableIPv4 and disableIPv6 flags).
*/
func (d *discoverer) TestHelperAddrs(
	ctx context.Context, thDomain string) (v4 string, v6 string, err error) {
	// 1. ensure the input domain name is valid
	if !domainRegexp.MatchString(thDomain) {
		return "", "", fmt.Errorf("invalid domain: %s", thDomain)
	}

	// 2. resolve the domain name to IP addrs
	reso := &dnscore.Resolver{}
	addrs, err := reso.LookupHost(ctx, thDomain)
	if err != nil {
		return "", "", err
	}

	// 3. find the first IPv4 address
	if !d.disableIPv4 {
		for _, addr := range addrs {
			ipaddr := netip.MustParseAddr(addr)
			if !ipaddr.Is4() {
				continue
			}
			v4 = addr
			break
		}
	}

	// 4. find the first IPv6 address
	if !d.disableIPv6 {
		for _, addr := range addrs {
			ipaddr := netip.MustParseAddr(addr)
			if !ipaddr.Is6() {
				continue
			}
			v6 = addr
			break
		}
	}

	// 5. handle the case where we did not find any addresses
	if v4 == "" && v6 == "" {
		err := fmt.Errorf("no suitable addresses found for domain: %s", thDomain)
		return "", "", err
	}

	// 6. return the addresses
	return v4, v6, nil
}
