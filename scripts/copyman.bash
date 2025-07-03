#!/bin/bash

# copymash.bash - copies manual pages from ./pkg/cli to ./docs/man
# SPDX-License-Identifier: GPL-3.0-or-later

set -euo pipefail

for srcname in $(find ./pkg/cli -type f -name README.md); do
	destname="./docs/man/rbmk-$(basename "$(dirname "$srcname")").md"
	cp -v "$srcname" "$destname"
done
mv -v "./docs/man/rbmk-rootcmd.md" "./docs/man/rbmk.md"

if [[ $# -gt 0 && $1 == "--fail-if-dirty" ]]; then
	[[ -z $(git status -s) ]]  # fail if dirty
fi
