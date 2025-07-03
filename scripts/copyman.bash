#!/bin/bash

# copyman.bash - copies manual pages and tutorials.
# SPDX-License-Identifier: GPL-3.0-or-later

set -euo pipefail

for srcname in $(find ./pkg/cli -type f -name README.md); do
	destname="./docs/man/rbmk-$(basename "$(dirname "$srcname")").md"
	cp -v "$srcname" "$destname"
done
mv -v "./docs/man/rbmk-rootcmd.md" "./docs/man/rbmk.md"

for name in $(find ./pkg/cli/tutorial -type f -name \*.md | grep -v 'README\.md'); do
	cp -v "$name" "./docs/tutorial/$(basename "$name")"
done

if [[ $# -gt 0 && $1 == "--fail-if-dirty" ]]; then
	[[ -z $(git status -s) ]]  # fail if dirty
fi
