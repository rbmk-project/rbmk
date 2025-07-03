// SPDX-License-Identifier: GPL-3.0-or-later

package markdown

import "github.com/rbmk-project/rbmk/pkg/common/cliutils"

// LazyMaybeRender returns a [cliutils.LazyHelpRenderer] that
// attempts to render the provide help string using markdown by
// calling [MaybeRender] when the help is requested.
func LazyMaybeRender(help string) cliutils.LazyHelpRenderer {
	return cliutils.LazyHelpRendererFunc(func() string {
		return MaybeRender(help)
	})
}
