// SPDX-License-Identifier: GPL-3.0-or-later

package netipx_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/rbmk-project/rbmk/pkg/common/netipx"
	"github.com/stretchr/testify/assert"
)

func TestAddrToAddrPort(t *testing.T) {
	tests := []struct {
		name string
		addr net.Addr
		want netip.AddrPort
	}{
		{
			name: "nil address",
			addr: nil,
			want: netip.AddrPortFrom(netip.IPv6Unspecified(), 0),
		},

		{
			name: "TCP address",
			addr: &net.TCPAddr{
				IP:   net.ParseIP("2001:db8::1"),
				Port: 1234,
			},
			want: netip.MustParseAddrPort("[2001:db8::1]:1234"),
		},

		{
			name: "UDP address",
			addr: &net.UDPAddr{
				IP:   net.ParseIP("2001:db8::2"),
				Port: 5678,
			},
			want: netip.MustParseAddrPort("[2001:db8::2]:5678"),
		},

		{
			name: "other address type",
			addr: &net.UnixAddr{},
			want: netip.AddrPortFrom(netip.IPv6Unspecified(), 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := netipx.AddrToAddrPort(tt.addr)
			assert.Equal(t, tt.want, got)
		})
	}
}
