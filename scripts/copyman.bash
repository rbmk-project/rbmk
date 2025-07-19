#!/bin/bash

# copyman.bash - copies manual pages and tutorials.
# SPDX-License-Identifier: GPL-3.0-or-later

set -euo pipefail

find ./pkg/cli -type f -name README.md -print0 | while IFS= read -r -d '' srcname; do
	destname="./docs/man/rbmk-$(basename "$(dirname "$srcname")").md"
	cp -v "$srcname" "$destname"
done
mv -v "./docs/man/rbmk-rootcmd.md" "./docs/man/rbmk.md"

find ./pkg/cli/tutorial -type f -name '*.md' ! -name 'README.md' -print0 | while IFS= read -r -d '' name; do
	cp -v "$name" "./docs/tutorial/$(basename "$name")"
done

if [[ $# -gt 0 && $1 == "--fail-if-dirty" ]]; then
	[[ -z $(git status -s) ]]  # fail if dirty
fi
