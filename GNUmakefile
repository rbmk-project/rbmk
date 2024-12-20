# SPDX-License-Identifier: GPL-3.0-or-later

#doc:
#doc: usage: make [target]
#doc:
#doc: We support the following targets:
.PHONY: help
help:
	@cat GNUmakefile | grep -E '^#doc:' | sed -e 's/^#doc: //g' -e 's/^#doc://'

#doc:
#doc: - `all`: builds `rbmk` and `rbmk-lite` for current platform
.PHONY: all
all: rbmk rbmk-lite

# Common variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/rbmk-project/rbmk/pkg/cli/version.Version=$(VERSION)
TAGS := netgo
GOENV := CGO_ENABLED=0
EXE ?=

#doc:
#doc: - `rbmk`: build rbmk in the current directory
#doc:
#doc: Use GOOS and GOARCH to force a specific architecture. For example, to
#doc: build for Linux on an ARM64 machine, run:
#doc:
#doc:     GOOS=linux GOARCH=arm64 make rbmk
#doc:
#doc: The resulting binary will be named `rbmk-linux-arm64-full`.
.PHONY: rbmk
rbmk:
	$(GOENV) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o rbmk-$(GOOS)-$(GOARCH)-full${EXE} -ldflags '$(LDFLAGS)' -tags $(TAGS) ./cmd/rbmk

#doc:
#doc: - `rbmk-lite`: same as `rbmk`, but with markdown rendering disabled
#doc: thus producing a significantly smaller binary.
#doc:
#doc: Use GOOS and GOARCH to force a specific architecture. For example, to
#doc: build for Linux on an ARM64 machine, run:
#doc:
#doc:     GOOS=linux GOARCH=arm64 make rbmk
#doc:
#doc: The resulting binary will be named `rbmk-linux-arm64-lite`.
.PHONY: rbmk-lite
rbmk-lite:
	$(GOENV) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o rbmk-$(GOOS)-$(GOARCH)-lite${EXE} -ldflags '$(LDFLAGS)' -tags $(TAGS),rbmk_disable_markdown ./cmd/rbmk

#doc:
#doc: - `release`: cross compile all platform/variant combinations
.PHONY: release
release:
	GOOS=linux GOARCH=amd64 $(MAKE) rbmk rbmk-lite
	GOOS=linux GOARCH=arm64 $(MAKE) rbmk rbmk-lite
	GOOS=windows GOARCH=amd64 EXE=.exe $(MAKE) rbmk rbmk-lite
	GOOS=darwin GOARCH=arm64 $(MAKE) rbmk rbmk-lite

#doc:
#doc: - `check`: run tests
.PHONY: check
check:
	go test -race -count 1 -cover ./...

#doc:
#doc: - `clean`: remove build artifacts
.PHONY: clean
clean:
	rm -f rbmk rbmk-lite rbmk-*


#doc:
#doc: - `install`: install rbmk into the system
#doc:
#doc: Installs the full version of rbmk for the current platform.
#doc: Use PREFIX to specify installation prefix (default: `/usr/local`).
#doc: For stated installations, use DESTDIR as usual.
#doc:
#doc: Examples:
#doc:     make install
#doc:     make PREFIX=/opt/rbmk install
#doc:     make DESTDIR=/tmp/stage PREFIX=/usr/local install
.PHONY: install
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

install: rbmk
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 rbmk-$(GOOS)-$(GOARCH)-full $(DESTDIR)$(BINDIR)/rbmk

#doc:
