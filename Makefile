# SPDX-License-Identifier: GPL-3.0-or-later

#doc:
#doc: usage: make [target]
#doc:
#doc: We support the following targets:
.PHONY: help
help:
	@cat Makefile | grep -E '^#doc:' | sed -e 's/^#doc: //g' -e 's/^#doc://'

#doc:
#doc: - `all`: alias for `rbmk`
.PHONY: all
all: rbmk

#doc:
#doc: - `rbmk`: build rbmk in the current directory
.PHONY: rbmk
rbmk:
	go build -v -ldflags '-s -w' -tags netgo ./cmd/rbmk

#doc:
#doc: - `stage`: prepare development environment in dist/
.PHONY: stage
stage: rbmk
	install -d dist/bin dist/libexec
	install -m755 rbmk dist/libexec/rbmk.real

#doc:
#doc: - `check`: run tests
.PHONY: check
check:
	go test -v ./...

#doc:
#doc: - `clean`: remove build artifacts
.PHONY: clean
clean:
	rm -f rbmk
	rm -rf dist/libexec/rbmk.real

#doc:
