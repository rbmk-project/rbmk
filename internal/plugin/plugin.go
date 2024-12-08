// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package plugin provides support for running RBMK plugins.

See XXX (design document URL).
*/
package plugin

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/rbmk-project/common/cliutils"
	"github.com/rbmk-project/rbmk/internal/markdown"
)

// LIBEXEC is the directory where RBMK plugins are installed.
var LIBEXEC = filepath.Join(".", "libexec")

//go:embed README.md
var readme string

// validPluginNameRegex is the regular expression that a plugin name must match.
var validPluginNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// resolvePluginBinary resolves the plugin binary name.
func resolvePluginBinary(name string) (string, error) {
	// TODO(bassosimone): honour RBMK_PLUGIN_PATH environment variable
	return makePluginBinary(LIBEXEC, name)
}

// errInvalidPluginName is the error returned when a plugin name is invalid.
var errInvalidPluginName = errors.New("invalid plugin name")

// makePluginBinary returns the full path to the plugin binary.
func makePluginBinary(dir, name string) (string, error) {
	if !validPluginNameRegex.MatchString(name) {
		return "", fmt.Errorf("%w: %s", errInvalidPluginName, name)
	}
	fullpath := filepath.Join(dir, "rbmk-plugin-"+name)
	stat, err := os.Stat(fullpath)
	if err != nil {
		return "", fmt.Errorf("no such plugin binary: %s", fullpath)
	}
	if !stat.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", fullpath)
	}
	// TODO(bassosimone): check that the file is executable
	return fullpath, nil
}

// NewCommand creates the `rbmk plugin` Command.
func NewCommand() cliutils.Command {
	return cliutils.NewCommandWithSubCommands(
		"plugin",
		markdown.LazyMaybeRender(readme),
		map[string]cliutils.Command{
			"run": newRunCommand(),
		},
	)
}
