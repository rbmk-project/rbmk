// SPDX-License-Identifier: GPL-3.0-or-later

// Package markdown contains code to optionally render markdown files.
//
// If you compile with `go build -tags rbmk_disable_markdown`, the [MaybeRender]
// function won't try to render the markdown content and will return the original
// content unmodified.
package markdown
