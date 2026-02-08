// SPDX-License-Identifier: GPL-3.0-or-later

package qacore_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/sud"
	"github.com/bassosimone/uis"
	"github.com/rbmk-project/rbmk/internal/qacore"
)

// Create the simulation once when the test code is run
var simulation = qacore.MustNewSimulation("testdata", qacore.ScenarioV4)

func Example() {
	// Create a background context
	ctx := context.Background()

	// Start a PCAP trace to capture all packets for debugging.
	// The resulting file can be inspected with Wireshark or tcpdump.
	filep := runtimex.PanicOnError1(os.Create("testdata/example.pcap"))
	pcaptrace := uis.NewPCAPTrace(filep, uis.MTUJumbo)
	defer pcaptrace.Close()

	// Set a packet filter that captures all packets to the trace
	simulation.SetPacketFilter(qacore.PacketFilterFunc(func(pkt uis.VNICFrame) bool {
		pcaptrace.Dump(pkt.Packet)
		return false
	}))

	// Resolve www.example.com
	addrs := runtimex.PanicOnError1(simulation.LookupHost(ctx, "www.example.com"))
	fmt.Printf("%+v\n", addrs)

	// Fetch www.example.com
	tcfg := &tls.Config{
		ServerName: "www.example.com",
		RootCAs:    simulation.CertPool(),
	}
	endpoint := net.JoinHostPort(addrs[0], "443")
	conn := runtimex.PanicOnError1(simulation.DialContext(ctx, "tcp", endpoint))
	defer conn.Close()

	tconn := tls.Client(conn, tcfg)
	defer tconn.Close()

	runtimex.PanicOnError0(tconn.HandshakeContext(ctx))
	suse := sud.NewSingleUseDialer(tconn)
	txp := &http.Transport{DialTLSContext: suse.DialContext}
	clnt := &http.Client{Transport: txp}
	hr := runtimex.PanicOnError1(clnt.Get("https://www.example.com/"))
	defer hr.Body.Close()

	fmt.Printf("%d\n", hr.StatusCode)

	body := runtimex.PanicOnError1(io.ReadAll(hr.Body))
	fmt.Printf("%d\n", len(body))

	// Output:
	// [104.18.26.120]
	// 200
	// 605
}
