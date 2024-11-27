// SPDX-License-Identifier: GPL-3.0-or-later

package qa

import "time"

// Driver is the interface for running QA scenarios. It is compatible with
// [testing.T] and testify's TestingT, allowing scenarios to be used both in
// automated tests and standalone QA
type Driver interface {
	// Deadline returns the suite deadline or false if there is no deadline.
	Deadline() (time.Time, bool)

	// Errorf formats and logs its arguments and records the error.
	Errorf(format string, args ...any)

	// FailNow immediately fails the current QA execution.
	FailNow()

	// Fatalf formats and logs its arguments and stop the execution of the suite.
	Fatalf(format string, args ...any)

	// Logf formats and logs the given message.
	Logf(format string, args ...any)
}
