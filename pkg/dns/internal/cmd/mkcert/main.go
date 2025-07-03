// SPDX-License-Identifier: GPL-3.0-or-later

// Command mkcert generates a self-signed certificate for testing purposes.
package main

import (
	"path/filepath"

	"github.com/rbmk-project/rbmk/pkg/common/selfsignedcert"
)

var destdir = filepath.Join("pkg", "dns", "dnscoretest")

func main() {
	cert := selfsignedcert.New(selfsignedcert.NewConfigExampleCom())
	cert.WriteFiles(destdir)
}
