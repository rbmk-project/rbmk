//
// SPDX-License-Identifier: BSD-3-Clause
//
// Adapted from: https://github.com/mvdan/sh
//
// Code to format shell scripts.
//

package generate

import (
	"bytes"
	"io"

	"mvdan.cc/sh/v3/syntax"
)

// shFormat formats the given shell script into the given [io.Writer].
func shFormatDump(code []byte, minify bool, w io.Writer) error {
	parser := syntax.NewParser(
		syntax.KeepComments(true),
		syntax.Variant(syntax.LangBash),
	)
	printer := syntax.NewPrinter(syntax.Minify(minify))
	node, err := parser.Parse(bytes.NewReader(code), "-")
	if err != nil {
		return err
	}
	syntax.Simplify(node)
	return printer.Print(w, node)
}
