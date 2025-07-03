// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package errclass implements error classification.

The general idea is to classify golang errors to an enum of strings
with names resembling standard Unix error names.

# Design Principles

1. Preserve original error in `err` in the structured logs.

2. Add the classified error as the `errClass` field.

3. Use [errors.Is] and [errors.As] for classification.

4. Use string-based classification for readability.

5. Follow Unix-like naming where appropriate.

6. Prefix subsystem-specific errors (`EDNS_`, `ETLS_`).

7. Keep full names for clarity over brevity.

8. Map the nil error to an empty string.

# System and Network Errors

- [ETIMEDOUT] for [context.DeadlineExceeded], [os.ErrDeadlineExceeded]

- [EINTR] for [context.Canceled], [net.ErrClosed]

- [EEOF] for (unexpected) [io.EOF] and [io.ErrUnexpectedEOF] errors

- [ECONNRESET], [ECONNREFUSED], ... for respective syscall errors

The actual system error constants are defined in platform-specific files:

- unix.go for Unix-like systems using x/sys/unix

- windows.go for Windows systems using x/sys/windows

This ensures proper mapping between the standardized error classes and
platform-specific error constants.

# DNS Errors

- [EDNS_NONAME] for errors with the "no such host" suffix

- [EDNS_NODATA] for errors with the "no answer" suffix

# TLS

- [ETLS_HOSTNAME_MISMATCH] for hostname verification failure

- [ETLS_CA_UNKNOWN] for unknown certificate authority

- [ETLS_CERT_INVALID] for invalid certificate

# Fallback

- [EGENERIC] for unclassified errors
*/
package errclass

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"os"
	"strings"
)

const (
	//
	// Errors that we can map using [errors.Is]:
	//

	// EADDRNOTAVAIL is the address not available error.
	EADDRNOTAVAIL = "EADDRNOTAVAIL"

	// EADDRINUSE is the address in use error.
	EADDRINUSE = "EADDRINUSE"

	// ECONNABORTED is the connection aborted error.
	ECONNABORTED = "ECONNABORTED"

	// ECONNREFUSED is the connection refused error.
	ECONNREFUSED = "ECONNREFUSED"

	// ECONNRESET is the connection reset by peer error.
	ECONNRESET = "ECONNRESET"

	// EHOSTUNREACH is the host unreachable error.
	EHOSTUNREACH = "EHOSTUNREACH"

	// EEOF indicates an unexpected EOF.
	EEOF = "EEOF"

	// EINVAL is the invalid argument error.
	EINVAL = "EINVAL"

	// EINTR is the interrupted system call error.
	EINTR = "EINTR"

	// ENETDOWN is the network is down error.
	ENETDOWN = "ENETDOWN"

	// ENETUNREACH is the network unreachable error.
	ENETUNREACH = "ENETUNREACH"

	// ENOBUFS is the no buffer space available error.
	ENOBUFS = "ENOBUFS"

	// ENOTCONN is the not connected error.
	ENOTCONN = "ENOTCONN"

	// EPROTONOSUPPORT is the protocol not supported error.
	EPROTONOSUPPORT = "EPROTONOSUPPORT"

	// ETIMEDOUT is the operation timed out error.
	ETIMEDOUT = "ETIMEDOUT"

	//
	// Errors that we can map using the error message suffix:
	//

	// EDNS_NONAME is the DNS error for "no such host".
	EDNS_NONAME = "EDNS_NONAME"

	// EDNS_NODATA 	is the DNS error for "no answer".
	EDNS_NODATA = "EDNS_NODATA"

	//
	// Errors that we can map using [errors.As]:
	//

	// ETLS_HOSTNAME_MISMATCH is the TLS error for hostname verification failure.
	ETLS_HOSTNAME_MISMATCH = "ETLS_HOSTNAME_MISMATCH"

	// ETLS_CA_UNKNOWN is the TLS error for unknown certificate authority.
	ETLS_CA_UNKNOWN = "ETLS_CA_UNKNOWN"

	// ETLS_CERT_INVALID is the TLS error for invalid certificate.
	ETLS_CERT_INVALID = "ETLS_CERT_INVALID"

	//
	// Fallback errors:
	//

	// EGENERIC is the generic, unclassified error.
	EGENERIC = "EGENERIC"
)

// errorsIsMap contains the errors that we can map with [errors.Is].
var errorsIsMap = map[error]string{
	context.DeadlineExceeded: ETIMEDOUT,
	context.Canceled:         EINTR,
	errEADDRNOTAVAIL:         EADDRNOTAVAIL,
	errEADDRINUSE:            EADDRINUSE,
	errECONNABORTED:          ECONNABORTED,
	errECONNREFUSED:          ECONNREFUSED,
	errECONNRESET:            ECONNRESET,
	errEHOSTUNREACH:          EHOSTUNREACH,
	io.EOF:                   EEOF,
	io.ErrUnexpectedEOF:      EEOF,
	errEINVAL:                EINVAL,
	errEINTR:                 EINTR,
	errENETDOWN:              ENETDOWN,
	errENETUNREACH:           ENETUNREACH,
	errENOBUFS:               ENOBUFS,
	errENOTCONN:              ENOTCONN,
	errEPROTONOSUPPORT:       EPROTONOSUPPORT,
	errETIMEDOUT:             ETIMEDOUT,
	net.ErrClosed:            EINTR,
	os.ErrDeadlineExceeded:   ETIMEDOUT,
}

// stringSuffixMap contains the errors that we can map using the error message suffix.
var stringSuffixMap = map[string]string{
	"no answer from DNS server": EDNS_NODATA,
	"no such host":              EDNS_NONAME,
}

// errorsAsList contains the errors that we can map with [errors.As].
var errorsAsList = []struct {
	as    func(err error) bool
	class string
}{
	{
		as: func(err error) bool {
			var candidate x509.HostnameError
			return errors.As(err, &candidate)
		},
		class: ETLS_HOSTNAME_MISMATCH,
	},

	{
		as: func(err error) bool {
			var candidate x509.UnknownAuthorityError
			return errors.As(err, &candidate)
		},
		class: ETLS_CA_UNKNOWN,
	},

	{
		as: func(err error) bool {
			var candidate x509.CertificateInvalidError
			return errors.As(err, &candidate)
		},
		class: ETLS_CERT_INVALID,
	},
}

// New creates a new error class from the given error.
func New(err error) string {
	// exclude the nil error case first
	if err == nil {
		return ""
	}

	// attemp direct mapping using the [errors.Is] func
	for candidate, class := range errorsIsMap {
		if errors.Is(err, candidate) {
			return class
		}
	}

	// attempt indirect mapping using the [errors.As] func
	for _, entry := range errorsAsList {
		if entry.as(err) {
			return entry.class
		}
	}

	// fallback to attempt matching with the string suffix
	for suffix, class := range stringSuffixMap {
		if strings.HasSuffix(err.Error(), suffix) {
			return class
		}
	}

	// we don't known this error
	return EGENERIC
}
