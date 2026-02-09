// SPDX-License-Identifier: GPL-3.0-or-later

package qacore_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/bassosimone/runtimex"
	"github.com/bassosimone/uis"
	"github.com/rbmk-project/rbmk/internal/qacore"
)

// Package-level simulation state
var (
	router    = qacore.NewDefaultRouter()
	simCtx    context.Context
	simCancel context.CancelFunc
)

// Create the simulation once when the test code is run
var simulation *qacore.Simulation

func init() {
	simCtx, simCancel = context.WithCancel(context.Background())
	simulation = qacore.MustNewSimulation(simCtx, "testdata", qacore.ScenarioV4(), router)
}

func TestMain(m *testing.M) {
	code := m.Run()
	simCancel()
	simulation.Wait()
	os.Exit(code)
}

func Example() {
	// Create a background context
	ctx := context.Background()

	// Start a PCAP trace to capture all packets for debugging.
	// The resulting file can be inspected with Wireshark or tcpdump.
	filep := runtimex.PanicOnError1(os.Create("testdata/example.pcap"))
	pcaptrace := uis.NewPCAPTrace(filep, uis.MTUJumbo)
	defer pcaptrace.Close()

	// Set a packet filter that captures all packets to the trace
	router.SetPacketFilter(qacore.PacketFilterFunc(func(pkt uis.VNICFrame) bool {
		pcaptrace.Dump(pkt.Packet)
		return false
	}))

	// Resolve www.example.com
	addrs := runtimex.PanicOnError1(simulation.LookupHost(ctx, "www.example.com"))
	fmt.Printf("%+v\n", addrs)

	// Fetch www.example.com
	txp := &http.Transport{
		DialContext:     simulation.DialContext,
		TLSClientConfig: &tls.Config{RootCAs: simulation.CertPool()},
	}
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
