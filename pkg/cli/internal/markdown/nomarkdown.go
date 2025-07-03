//go:build rbmk_disable_markdown

// SPDX-License-Identifier: GPL-3.0-or-later

package markdown

// MaybeRender tries to render the given markdown content. On error,
// it returns the original unmodified content.
func MaybeRender(content string) string {
	return content
}
