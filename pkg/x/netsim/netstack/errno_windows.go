//go:build windows

//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Windows errno definitions.
//

package netstack

import "golang.org/x/sys/windows"

const (
	// EADDRNOTAVAIL is the address not available error.
	EADDRNOTAVAIL = windows.WSAEADDRNOTAVAIL

	// EADDRINUSE is the address in use error.
	EADDRINUSE = windows.WSAEADDRINUSE

	// ECONNABORTED is the connection aborted error.
	ECONNABORTED = windows.WSAECONNABORTED

	// ECONNREFUSED is the connection refused error.
	ECONNREFUSED = windows.WSAECONNREFUSED

	// ECONNRESET is the connection reset by peer error.
	ECONNRESET = windows.WSAECONNRESET

	// EHOSTUNREACH is the host unreachable error.
	EHOSTUNREACH = windows.WSAEHOSTUNREACH

	// EINVAL is the invalid argument error.
	EINVAL = windows.WSAEINVAL

	// ENETDOWN is the network is down error.
	ENETDOWN = windows.WSAENETDOWN

	// ENOBUFS is the no buffer space available error.
	ENOBUFS = windows.WSAENOBUFS

	// ENOTCONN is the not connected error.
	ENOTCONN = windows.WSAENOTCONN

	// EPROTONOSUPPORT is the protocol not supported error.
	EPROTONOSUPPORT = windows.WSAEPROTONOSUPPORT
)
