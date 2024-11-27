// SPDX-License-Identifier: GPL-3.0-or-later

package qa_test

import (
	"testing"

	"github.com/rbmk-project/rbmk/internal/qa"
)

func TestQA(t *testing.T) {
	for _, scenario := range qa.Registry {
		t.Run(scenario.Name, func(t *testing.T) {
			scenario.Run(t)
		})
	}
}
