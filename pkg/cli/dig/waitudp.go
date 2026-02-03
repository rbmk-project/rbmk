// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"log/slog"
	"net/netip"

	"github.com/bassosimone/dnscodec"
	"github.com/rbmk-project/rbmk/pkg/common/closepool"
)

// waitUDP sends a query over UDP and waits for duplicate responses
// until the timeout configured in the context expires.
func (task *Task) waitUDP(
	ctx context.Context,
	cp *closepool.Pool,
	slogger *slog.Logger,
	serverAddr netip.AddrPort,
	query *dnscodec.Query,
) error {
	// TODO(bassosimone): we probably need to implement support
	// for this inside the `nop` package. We need to implement
	// support for manually sending a query and receiving a response
	panic("not implemented")
}
