//go:build !rbmk_disable_plugin

// SPDX-License-Identifier: GPL-3.0-or-later

package plugin

import "github.com/rbmk-project/rbmk/internal/plugin"

var newCommand = plugin.NewCommand
