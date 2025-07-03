// SPDX-License-Identifier: GPL-3.0-or-later

package rbmk_test

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// packageGroup describes a group of packages and their allowed dependencies.
type packageGroup struct {
	// Name is the name of the package group.
	Name string

	// Allowed is a list of allowed dependencies for the package group.
	Allowed []string
}

// groups lists all known package groups.
var groups = []packageGroup{
	// pkg/cli can depend on basically everything.
	{
		Name: "pkg/cli",
		Allowed: []string{
			"internal",
			"pkg/cli",
			"pkg/common",
			"pkg/dns",
			"pkg/x",
		},
	},

	// pkg/common can only depend on itself.
	{
		Name: "pkg/common",
		Allowed: []string{
			"pkg/common",
		},
	},

	// pkg/dns can depend on pkg/common and itself.
	{
		Name: "pkg/dns",
		Allowed: []string{
			"pkg/common",
			"pkg/dns",
		},
	},

	// pkg/x can depend on pkg/common, pkg/dns, and itself.
	{
		Name: "pkg/x",
		Allowed: []string{
			"pkg/common",
			"pkg/dns",
			"pkg/x",
		},
	},
}

// validateSpecificGroup validates a specific package group against its allowed dependencies.
func validateSpecificGroup(t *testing.T, modpath string, group packageGroup) {
	// Make an allow list containing fully qualified package names
	allow := make([]string, 0, len(group.Allowed))
	for _, entry := range group.Allowed {
		allow = append(allow, modpath+"/"+entry)
	}

	// Load all packages in the group
	config := &packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps}
	fullname := modpath + "/" + group.Name + "/..."
	pkgs, err := packages.Load(config, fullname)
	if err != nil {
		t.Errorf("error loading %q: %s", fullname, err.Error())
		return
	}

	// Process each loaded package
	for _, pkg := range pkgs {

		// Process each import used by the package
		for _, dep := range pkg.Imports {

			// Skip dependencies outside of the module prefix
			if !strings.HasPrefix(dep.PkgPath, modpath) {
				continue
			}

			// Ensure the dependency is allowed
			var found bool
			for _, entry := range allow {
				found = found || strings.HasPrefix(dep.PkgPath, entry)
			}
			if !found {
				t.Errorf("package %q depends on %q, which is not listed in %v", pkg.PkgPath, dep.PkgPath, allow)
				continue
			}
		}
	}
}

// validateAllGroups validates all package groups against their allowed dependencies.
func validateAllGroups(t *testing.T, modpath string, groups []packageGroup) {
	for _, group := range groups {
		t.Run(group.Name, func(t *testing.T) {
			validateSpecificGroup(t, modpath, group)
		})
	}
}

// validateGroupNames ensures that the group names listed in groups are
// consistent with the package dirs inside the `./pkg` directory.
func validateGroupNames(t *testing.T, groups []packageGroup) {
	dentries, err := os.ReadDir("pkg")
	if err != nil {
		t.Fatalf("error reading package directory: %v", err)
		return
	}

	const (
		actual = 1 << iota
		specced
	)
	accounted := make(map[string]int, len(dentries))
	for _, dentry := range dentries {
		if dentry.IsDir() && dentry.Name() != "testdata" {
			accounted[dentry.Name()] |= actual
		}
	}
	for _, group := range groups {
		accounted[strings.TrimPrefix(group.Name, "pkg/")] |= specced
	}

	for name, flags := range accounted {
		switch flags {
		case actual | specced:
			// all good
		case actual:
			t.Errorf("package group %q is not listed in the package specifications but has a corresponding directory", name)
		default:
			t.Errorf("package group %q is listed in the package specifications but does not have a corresponding directory", name)
		}
	}
}

func TestPublicPackages(t *testing.T) {
	validateGroupNames(t, groups)
	validateAllGroups(t, "github.com/rbmk-project/rbmk", groups)
}
