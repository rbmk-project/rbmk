// SPDX-License-Identifier: GPL-3.0-or-later

package qa

// Registry is the list of all the available [ScenarioDescriptor].
var Registry = []ScenarioDescriptor{
	{
		Name:    "dnsOverUdpSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},

	{
		Name: "dnsOverUdpCensorship",
		Editors: []ScenarioEditor{
			CensorDNSLikeIran("www.example.com"),
		},
		Argv: []string{
			"rbmk", "dig", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},
}
