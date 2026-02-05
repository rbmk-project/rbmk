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
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
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
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
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
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
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
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
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
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "tlsHandshakeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsQuery"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "dnsResponse"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnyClose},
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
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "tlsHandshakeDone"},
			{Msg: "dnsExchangeStart"},
			{Msg: "dnsQuery"},
			{Msg: "httpRoundTripStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "httpRoundTripDone"},
			{Msg: "httpBodyStreamStart"},
			{Pattern: MatchAnyRead | MatchAnyWrite},
			{Msg: "httpBodyStreamDone"},
			{Msg: "dnsResponse"},
			{Msg: "dnsExchangeDone"},
			{Pattern: MatchAnyRead | MatchAnyWrite | MatchAnyClose},
		},
	},
}
