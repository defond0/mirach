PHONY: all archives clean clean-build deploy-docs docs help lint test
.DEFAULT_GOAL := help

# project variables
PROJECT_NAME := mirach
VERSION := $(shell git describe --always --dirty)

# helper variables
BUILDDIR := ./_build
ARCDIR := $(BUILDDIR)/arc
BINDIR := $(BUILDDIR)/bin
LDFLAGS := "-X main.version=$(VERSION)"

help:
	$(info available targets:)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		nb = sub( /^## /, "", helpMsg ); \
		if(nb == 0) { \
			helpMsg = $$0; \
			nb = sub( /^[^:]*:.* ## /, "", helpMsg ); \
		} \
		if (nb) \
			print  $$1 "\t" helpMsg; \
	} \
	{ helpMsg = $$0 }' \
	$(MAKEFILE_LIST) | column -ts $$'\t' | \
	grep --color '^[^ ]*'

SYSTEMS := linux windows
ARCHS := 386 amd64

define PROGRAM_template
PROG_TARGETS += $(BINDIR)/$(PROJECT_NAME)_$(VERSION)_$(1)_$(2)/$(PROJECT_NAME)
$(BINDIR)/$(PROJECT_NAME)_$(1)_$(2)/$(PROJECT_NAME): export GOOS = $(1)
$(BINDIR)/$(PROJECT_NAME)_$(1)_$(2)/$(PROJECT_NAME): export GOARCH = $(2)
ARC_TARGETS += $(ARCDIR)/$(PROJECT_NAME)_$(VERSION)_$(1)_$(2).zip
endef

$(foreach sys,$(SYSTEMS),$(foreach arch,$(ARCHS),$(eval $(call PROGRAM_template,$(sys),$(arch)))))

$(PROG_TARGETS):
	go build -i -v -ldflags=$(LDFLAGS) -o $@

$(ARCDIR)/%.zip: $(BINDIR)/%/*
	@mkdir -p $(ARCDIR)
	zip -j $@ $<

all: test $(PROG_TARGETS) archives ## build all systems and architectures

archives: $(ARC_TARGETS) ## archive all builds

clean: clean-build ## clean all

clean-build: ## remove build artifacts
	rm -rf $(BUILDDIR)

deploy-docs: docs ## deploy docs to S3 bucket
	aws s3 sync ./docs/html s3://***REMOVED***/$(PROJECT_NAME)/

docs: ## generate docs
	godoc . > ./docs/html/$(PROJECT_NAME).html

install: ## install to GOPATH
	go install -v -ldflags=$(LDFLAGS)

lint: ## gofmt goimports
	gofmt *.go
	-goimport *.go

publish:
	@echo "push to s3 at some point"

release: req-release-type req-release-repo clean ## package and upload a release
	release -t $(RELEASE_TYPE) -g $(RELEASE_REPO) $(RELEASE_BRANCH) $(RELEASE_BASE)

req-release-type:
ifndef RELEASE_TYPE
	$(error RELEASE_TYPE is undefined)
endif

req-release-repo:
ifndef RELEASE_REPO
	$(error RELEASE_REPO is undefined)
endif

test: ## run unit tests
	@echo "No unit tests for now, but make test-integration will run integration tests"

test-integration: ## run integration tests
	go build -race .
	go test -v -tags=integration .
