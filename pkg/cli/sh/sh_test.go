// SPDX-License-Identifier: GPL-3.0-or-later

package sh

import (
	"os"
	"testing"
)

// Ensure that `$RBMK_EXE` expands to `rbmk` in the common case.
func Test_osGetenvWrapper(t *testing.T) {
	// Make sure the variable is not set already
	if os.Getenv(rbmkExeVarName) != "" {
		t.Skip("the environment variable is already set")
	}

	// Check whether we get the expected value (yes, there is a race
	// with a potential os.Setenv in this test but we're not going
	// to call os.Setenv here so we're good).
	if value := osGetenvWrapper(rbmkExeVarName); value != "rbmk" {
		t.Error("unexpected value", value)
	}
}
