// SPDX-License-Identifier: GPL-3.0-or-later

package main

import "testing"

func Test_main(t *testing.T) {
	mainArgs = []string{"rbmk", "help"}
	main()
}
