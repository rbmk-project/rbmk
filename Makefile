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
#doc: - `check`: run tests
.PHONY: check
check:
	go test -race -count 1 -cover ./...

#doc:
#doc: - `clean`: remove build artifacts
.PHONY: clean
clean:
	rm -f rbmk

#doc:
