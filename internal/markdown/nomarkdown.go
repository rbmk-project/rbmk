//go:build rbmk_disable_markdown

// SPDX-License-Identifier: GPL-3.0-or-later

package markdown

// TryRender tries to render the given markdown content. On error,
// it returns the original unmodified content.
func TryRender(content string) string {
	return content
}
