// SPDX-License-Identifier: GPL-3.0-or-later

package errclass

import (
	"crypto/x509"
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	// testcase is a test case implemented by this function.
	type testcase struct {
		input  error
		expect string
	}

	// start with a test case for the nil error
	var tests = []testcase{
		{
			input:  nil,
			expect: "",
		},
	}

	// add tests for cases we can test with errors.Is
	for key, value := range errorsIsMap {
		tests = append(tests, testcase{
			input:  key,
			expect: value,
		})
	}

	// add tests for cases we can test with string suffix matching
	for suffix, class := range stringSuffixMap {
		tests = append(tests, testcase{
			input:  errors.New("some error message " + suffix),
			expect: class,
		})
	}

	// add tests for cases we can test with errors.As
	tests = append(tests, testcase{
		input: x509.HostnameError{
			Certificate: &x509.Certificate{},
			Host:        "",
		},
		expect: ETLS_HOSTNAME_MISMATCH,
	})
	tests = append(tests, testcase{
		input: x509.UnknownAuthorityError{
			Cert: &x509.Certificate{},
		},
		expect: ETLS_CA_UNKNOWN,
	})
	tests = append(tests, testcase{
		input: x509.CertificateInvalidError{
			Cert:   &x509.Certificate{},
			Reason: 0,
			Detail: "",
		},
		expect: ETLS_CERT_INVALID,
	})

	// add test for unknown error
	tests = append(tests, testcase{
		input:  errors.New("unknown error"),
		expect: EGENERIC,
	})

	// run all tests
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			got := New(tt.input)
			if got != tt.expect {
				t.Errorf("New(%v) = %v; want %v", tt.input, got, tt.expect)
			}
		})
	}
}
