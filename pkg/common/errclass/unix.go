//go:build unix

// SPDX-License-Identifier: GPL-3.0-or-later

package errclass

import "golang.org/x/sys/unix"

const (
	errEADDRNOTAVAIL   = unix.EADDRNOTAVAIL
	errEADDRINUSE      = unix.EADDRINUSE
	errECONNABORTED    = unix.ECONNABORTED
	errECONNREFUSED    = unix.ECONNREFUSED
	errECONNRESET      = unix.ECONNRESET
	errEHOSTUNREACH    = unix.EHOSTUNREACH
	errEINVAL          = unix.EINVAL
	errEINTR           = unix.EINTR
	errENETDOWN        = unix.ENETDOWN
	errENETUNREACH     = unix.ENETUNREACH
	errENOBUFS         = unix.ENOBUFS
	errENOTCONN        = unix.ENOTCONN
	errEPROTONOSUPPORT = unix.EPROTONOSUPPORT
	errETIMEDOUT       = unix.ETIMEDOUT
)
