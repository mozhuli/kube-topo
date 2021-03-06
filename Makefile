# stackube Makefile
# Follows the interface defined in the Golang CTI proposed
# in https://review.openstack.org/410355

#REPO_VERSION?=$(shell git describe --tags)

GIT_HOST = git.openstack.org
SHELL := /bin/bash

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))
# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
PKG := git.openstack.org/openstack/stackube
DEST := $(GOPATH)/src/$(PKG)

GOFLAGS :=
TAGS :=
LDFLAGS :=

OUTPUT := _output

# Default target
.PHONY: all
all: build

# CTI targets

.PHONY: depend
depend: work
	# cd $(DEST) && glide install

.PHONY: depend-update
depend-update: work
	# cd $(DEST) && glide update

.PHONY: build
build: depend
	cd $(DEST)
	go build $(GOFLAGS) -a -o $(OUTPUT)/kube-topo ./cmd/kube-topo

.PHONY: install
install: depend
	cd $(DEST)
	install -D -m 755 $(OUTPUT)/kube-topo /usr/local/bin/kube-topo
	
.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit: depend
test-unit: TAGS += unit
test-unit: test-flags

.PHONY: test-flags
test-flags:
	cd $(DEST) && go test $(GOFLAGS) -tags '$(TAGS)' $(go list ./... | grep -v vendor)

# The above pipeline is required because gofmt always returns 0 and we need
# to detect if any files are listed as having format problems.
.PHONY: fmt
fmt: work
	files=$$(cd $(DEST) && find . -not \(  \( -wholename '*/vendor/*' \) -prune \) -name '*.go' | xargs gofmt -s -l | tee >(cat - >&2)); [ -z "$$files" ]

.PHONY: fmtfix
fmtfix: work
	cd $(DEST) && go fmt ./...

lint:
	@echo "$@ not yet implemented"

cover:
	@echo "$@ not yet implemented"

docs:
	@echo "$@ not yet implemented"

godoc:
	@echo "$@ not yet implemented"

releasenotes:
	@echo "Reno not yet implemented for this repo"

translation:
	@echo "$@ not yet implemented"

# Do the work here

# Set up the development environment
env:
	@echo "PWD: $(PWD)"
	@echo "BASE_DIR: $(BASE_DIR)"
	@echo "GOPATH: $(GOPATH)"
	@echo "DEST: $(DEST)"
	@echo "PKG: $(PKG)"

# Get our dev/test dependencies in place
bootstrap:
	tools/test-setup.sh

work: $(GOPATH) $(DEST)

$(GOPATH):
	mkdir -p $(GOPATH)

$(DEST): $(GOPATH)
	mkdir -p $(shell dirname $(DEST))
	ln -s $(PWD) $(DEST)

.bindep:
	virtualenv .bindep
	.bindep/bin/pip install bindep

bindep: .bindep
	@.bindep/bin/bindep -b -f bindep.txt || true

install-distro-packages:
	tools/install-distro-packages.sh

clean:
	rm -rf .bindep

realclean: clean
	rm -rf vendor
	if [ "$(GOPATH)" = "$(GOPATH_DEFAULT)" ]; then \
		rm -rf $(GOPATH); \
	fi

shell: work
	cd $(DEST) && $(SHELL) -i

.PHONY: bindep clean cover depend docs fmt functional lint realclean \
	relnotes test translation
