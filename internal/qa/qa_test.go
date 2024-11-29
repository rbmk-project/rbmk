// SPDX-License-Identifier: GPL-3.0-or-later

package qa_test

import (
	"testing"

	"github.com/rbmk-project/rbmk/internal/qa"
)

func TestQA(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	for _, scenario := range qa.Registry {
		t.Run(scenario.Name, func(t *testing.T) {
			scenario.VerifyEvents(t, scenario.Run(t))
		})
	}
}
