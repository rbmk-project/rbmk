//go:build !rbmk_disable_markdown

// SPDX-License-Identifier: GPL-3.0-or-later

// Package markdown contains code to render markdown files.
package markdown

import "github.com/charmbracelet/glamour"

// MaybeRender tries to render the given markdown content. On error,
// it returns the original unmodified content.
func MaybeRender(content string) string {
	render, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithPreservedNewLines())
	if err != nil {
		return content
	}
	out, err := render.Render(content)
	if err != nil {
		return content
	}
	return out
}
