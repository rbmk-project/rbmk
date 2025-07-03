// SPDX-License-Identifier: GPL-3.0-or-later

package netcore_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/rbmk-project/rbmk/pkg/x/netcore"
)

func TestDialerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	netx := &netcore.Network{}
	netx.Logger = logger
	netx.WrapConn = netcore.WrapConn

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := netx.DialContext(ctx, "tcp", "example.com:80")
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
}

func TestTLSDialerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	netx := &netcore.Network{}
	netx.Logger = logger
	netx.WrapConn = netcore.WrapConn

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := netx.DialTLSContext(ctx, "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
}
