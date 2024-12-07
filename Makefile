# SPDX-License-Identifier: GPL-3.0-or-later

#doc:
#doc: usage: make [target]
#doc:
#doc: We support the following targets:
.PHONY: help
help:
	@cat Makefile | grep -E '^#doc:' | sed -e 's/^#doc: //g' -e 's/^#doc://'

#doc:
#doc: - `all`: builds `rbmk` and `rbmk-small`
.PHONY: all
all: rbmk rbmk-small

#doc:
#doc: - `rbmk`: build rbmk in the current directory
.PHONY: rbmk
rbmk:
	go build -v -o rbmk -ldflags '-s -w' -tags netgo ./cmd/rbmk

#doc:
#doc: - `rbmk-small`: build rbmk without optional features in the current dir
.PHONY: rbmk-small
rbmk-small:
	go build -v -o rbmk-small -ldflags '-s -w' -tags rbmk_disable_markdown,netgo ./cmd/rbmk

#doc:
#doc: - `check`: run tests
.PHONY: check
check:
	go test -race -count 1 -cover ./...

#doc:
#doc: - `clean`: remove build artifacts
.PHONY: clean
clean:
	rm -f rbmk rbmk-small

#doc:
