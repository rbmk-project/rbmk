// SPDX-License-Identifier: GPL-3.0-or-later

package stun

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/pion/stun/v3"
	"github.com/rbmk-project/rbmk/internal/testable"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
	"github.com/rbmk-project/rbmk/pkg/x/netcore"
)

// Task runs a STUN binding request.
type Task struct {
	// Endpoint is the STUN server endpoint (HOST:PORT)
	Endpoint string

	// LogsWriter is where we write structured logs
	LogsWriter io.Writer

	// MaxTime is the maximum time to wait for the operation to finish.
	MaxTime time.Duration

	// Output is where we write the results
	Output io.Writer
}

// Run executes the STUN binding request task
func (task *Task) Run(ctx context.Context) error {
	// 1. Set up the overall operation timeout
	ctx, cancel := context.WithTimeout(ctx, task.MaxTime)
	defer cancel()

	// 2. Set up the JSON logger for writing measurements
	logger := slog.New(slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{}))

	// 3. Create a pool containing closers
	pool := &closepool.Pool{}
	defer pool.Close()

	// 4. Create netcore network instance
	netx := &netcore.Network{}
	netx.DialContextFunc = testable.DialContext.Get()
	netx.Logger = logger
	netx.WrapConn = func(ctx context.Context, netx *netcore.Network, conn net.Conn) net.Conn {
		conn = netcore.WrapConn(ctx, netx, conn)
		pool.Add(conn)
		return conn
	}

	// 5. Establish UDP connection to STUN server and make sure
	// we have proper context deadline propagation. Also, make sure
	// that we bail immediately if the context is done.
	conn, err := netx.DialContext(ctx, "udp", task.Endpoint)
	if err != nil {
		return fmt.Errorf("cannot connect to STUN server: %w", err)
	}
	if d, ok := ctx.Deadline(); ok {
		conn.SetDeadline(d)
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// 6. Build STUN binding request message
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// TODO(bassosimone): log the raw STUN request

	// 7. Send the STUN request
	if _, err := conn.Write(message.Raw); err != nil {
		return fmt.Errorf("cannot send STUN request: %w", err)
	}

	// 8. Read the response
	buffer := make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("cannot read STUN response: %w", err)
	}

	// 9. Parse the STUN response
	response := &stun.Message{Raw: buffer[:count]}
	if err := response.Decode(); err != nil {
		return fmt.Errorf("cannot decode STUN response: %w", err)
	}

	// TODO(bassosimone): log the raw STUN response

	// 10. Extract the XOR-MAPPED-ADDRESS
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(response); err != nil {
		return fmt.Errorf("cannot get reflexive address: %w", err)
	}

	// 11. Log and print the reflexive address
	fmt.Fprintf(task.Output, "%s\n", net.JoinHostPort(
		xorAddr.IP.String(), strconv.Itoa(xorAddr.Port)))
	logger.InfoContext(
		ctx,
		"stunReflexiveAddress",
		"stunReflexiveIPAddr", xorAddr.IP.String(),
		"stunReflexivePort", xorAddr.Port,
	)

	// 12. Explicitly close connections in the pool
	pool.Close()
	return nil
}
