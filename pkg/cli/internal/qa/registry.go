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
			"rbmk", "dig", "+noall", "+logs", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyClose},
		},
	},

	{
		Name: "dnsOverUdpCensorship",
		Editors: []ScenarioEditor{
			CensorDNSLikeIran("www.example.com"),
		},
		Argv: []string{
			"rbmk", "dig", "+noall", "+logs", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyClose},
		},
	},

	{
		Name: "dnsOverUdpCensorshipWithDuplicates",
		Editors: []ScenarioEditor{
			CensorDNSLikeIran("www.example.com"),
		},
		Argv: []string{
			"rbmk", "dig", "+udp=wait-duplicates", "+noall", "+logs", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyClose},
		},
	},

	//
	// DNS over TCP
	//

	{
		Name:    "dnsOverTcpSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+noall", "+logs", "+tcp", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyClose},
		},
	},

	//
	// DNS over TLS
	//

	{
		Name:    "dnsOverTlsSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+noall", "+logs", "+tls", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "tlsHandshakeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "tlsHandshakeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline | MatchAnyClose},
		},
	},

	//
	// DNS over HTTPS
	//

	{
		Name:    "dnsOverHttpsSuccess",
		Editors: []ScenarioEditor{},
		Argv: []string{
			"rbmk", "dig", "+noall", "+logs", "+https", "@8.8.8.8", "A", "www.example.com",
		},
		ExpectedErr: nil,
		ExpectedSeq: []ExpectedEvent{
			{Msg: "connectStart"},
			{Msg: "connectDone"},
			{Msg: "tlsHandshakeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "tlsHandshakeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "httpRoundTripStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "httpRoundTripDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "httpBodyStreamStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "httpBodyStreamDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnySetDeadline},
		},
	},
}
