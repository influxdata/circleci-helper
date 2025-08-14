ARCH := $(shell uname -m | sed 's/x86_64/amd64_v1/')
GOFILES := $(wildcard cmd/circleci-helper/*.go) $(wildcard cmd/circleci-helper/**/*.go) $(wildcard cmd/circleci-helper/**/**/*.go)
GORELEASER_VERSION := v1.25.1
OS := $(shell uname -o | tr [:upper:] [:lower:])

.PHONY: all
all: circleci-helper

.PHONY: help
help: ## print this message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\-\%]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: debug-%
debug-%: ## debug make variables (try make debug-GOFILES for an example)
	@echo $(*) = $($(*))

.PHONY: build
build: circleci-helper ## alias of circleci-helper

build-all: tmp/go/bin/goreleaser $(GOFILES) ## build circleci-helper for all supported arches / OSes
	tmp/go/bin/goreleaser build --clean --snapshot

circleci-helper: dist/circleci-helper_$(OS)_$(ARCH)/circleci-helper ## build circleci-helper for current arch / OS
	cp $< .

.PHONY: clean
clean: ## clean ./circleci-helper & dist/
	-rm -r circleci-helper dist/

.PHONY: clean-all
clean-all: clean ## clean + goreleaser binary
	-rm -r tmp/

dist/circleci-helper_%/circleci-helper: tmp/go/bin/goreleaser $(GO_FILES)
	GOOS=$(shell echo $(@) | cut -f 2 -d / | cut -f 2 -d _) \
	GOARCH=$(shell echo $(@) | cut -f 2 -d / | cut -f 3- -d _) \
	tmp/go/bin/goreleaser build --clean --single-target --snapshot

.SILENT: tmp/go/bin/goreleaser
tmp/go/bin/goreleaser:
	echo installing goreleaser to tmp/go/bin/goreleaser...
	GOPATH=$(PWD)/tmp/go go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) 2>/dev/null
	-chmod -R 700 tmp/go/pkg
	-rm -rf tmp/go/pkg
