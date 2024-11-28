// SPDX-License-Identifier: GPL-3.0-or-later

package qa

// Registry is the list of all the available [ScenarioDescriptor].
var Registry = []ScenarioDescriptor{

	//
	// DNS over UDP
	//

	{
		Name:    "dnsOverUdpSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+logs", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},

	{
		Name: "dnsOverUdpCensorship",
		Editors: []ScenarioEditor{
			CensorDNSLikeIran("www.example.com"),
		},
		Argv: []string{
			"rbmk", "dig", "+logs", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},

	//
	// DNS over TCP
	//

	{
		Name:    "dnsOverTcpSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+logs", "+tcp", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},

	//
	// DNS over TLS
	//

	{
		Name:    "dnsOverTlsSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+logs", "+tls", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},

	//
	// DNS over HTTPS
	//

	{
		Name:    "dnsOverHttpsSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+logs", "+https", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
	},
}
