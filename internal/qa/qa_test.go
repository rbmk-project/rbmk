// SPDX-License-Identifier: GPL-3.0-or-later

package qa_test

import (
	"io"
	"testing"

	"github.com/rbmk-project/common/runtimex"
	"github.com/rbmk-project/rbmk/internal/qa"
)

func TestQA(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	for _, scenario := range qa.Registry {
		t.Run(scenario.Name, func(t *testing.T) {
			logsReader := scenario.Run(t)

			// TODO(bassosimone):
			//
			// 1. implement `./cmd/qatool` to generate golden files
			//
			// 2. implement checking the logs against the golden files
			t.Log(string(runtimex.Try1(io.ReadAll(logsReader))))
		})
	}
}
