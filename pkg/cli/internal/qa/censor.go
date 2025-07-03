// SPDX-License-Identifier: GPL-3.0-or-later

package qa

import (
	"github.com/rbmk-project/rbmk/pkg/x/netsim"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/censor"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/dns"
)

// CensorDNSLikeIran returns a ScenarioEditor that implements Iran-like
// DNS censorship for the given domains. When applied, DNS queries for
// these domains will receive poisoned responses pointing to known
// Iranian block pages (10.10.34.34-36).
func CensorDNSLikeIran(domains ...string) ScenarioEditor {
	return func(scenario *netsim.Scenario) *netsim.Scenario {
		ddb := dns.NewDatabase()
		ddb.AddAddresses(domains, []string{
			"10.10.34.34",
			"10.10.34.35",
			"10.10.34.36",
		})
		scenario.Router().AddFilter(censor.NewDNSPoisoner(ddb))
		return scenario
	}
}
