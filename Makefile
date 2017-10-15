PHONY: all all-snap archives clean clean-build deploy-docs docs help install install-build-deps install-test-deps lint publish publish-release publish-snapshot release test test-all test-integration
.DEFAULT_GOAL := help

# project variables
PROJECT_NAME := mirach
VERSION := $(if $(SNAP),SNAPSHOT,$(shell git describe --always --dirty))

# helper variables
BUILDDIR := ./_build
ARCDIR := $(BUILDDIR)/arc
BINDIR := $(BUILDDIR)/bin
COV := cover.out
DOWNLOADLOC := s3://mirach/builds
DOWNLOADSNAPLOC := $(DOWNLOADLOC)/SNAPSHOT
DOWNLOADSRELEASELOC := $(DOWNLOADLOC)/RELEASE
ROOTPKG := github.com/cleardataeng/mirach
LDFLAGS := "-X $(ROOTPKG)/util.Version=$(VERSION)"

help:
	$(info available targets:)
	@awk '/^[a-zA-Z\-\_0-9\.]+:/ { \
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

PROG_TARGETS :=
define PROGRAM_template
CUR := $(BINDIR)/$(PROJECT_NAME)_$(VERSION)_$(1)_$(2)/$(PROJECT_NAME)$(if $(filter windows,$(1)),.exe)
$$(CUR): export GOOS = $(1)
$$(CUR): export GOARCH = $(2)
PROG_TARGETS += $$(CUR)
ARC_TARGETS += $(ARCDIR)/$(PROJECT_NAME)_$(VERSION)_$(1)_$(2).zip
endef

$(foreach sys,$(SYSTEMS),$(foreach arch,$(ARCHS),$(eval $(call PROGRAM_template,$(sys),$(arch)))))

$(PROG_TARGETS):
	go build -i -v -ldflags=$(LDFLAGS) -o $@

$(ARCDIR)/%.zip: $(BINDIR)/%/*
	@mkdir -p $(ARCDIR)
	zip -j $@ $<

all: test $(PROG_TARGETS) archives ## build all systems and architectures

all-snap:
	SNAP="true" $(MAKE) all

archives: $(ARC_TARGETS) ## archive all builds

clean: clean-build ## clean all

clean-build: ## remove build artifacts
	rm -rf $(BUILDDIR)

clean-coverage: ## remove coverage profile
	rm -f $(COV)

coverage-browser: $(COV) ## open coverage report in browser
	go tool cover -html=$(COV)

coverage-deps: ## installs gocoverutil for checking coverage of full lib
	go get -u github.com/AlekSi/gocoverutil

install: install-build-deps ## install to GOPATH
	go install -v -ldflags=$(LDFLAGS)

install-build-deps: ## install go dependencies
	go get ./...

lint: ## gofmt goimports
	gofmt *.go
	-goimport *.go

publish: ## publish all current build archives
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADLOC)"
	aws s3 sync $(ARCDIR)/ $(DOWNLOADLOC)/

publish-release: publish ## publish current build archives to release location
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADSRELEASELOC)"
	aws s3 sync --delete $(ARCDIR)/ $(DOWNLOADSRELEASELOC)/

publish-snapshot: ## publish current build archives to snapshot location
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADSNAPLOC)"
	aws s3 sync --delete $(ARCDIR)/ $(DOWNLOADSNAPLOC)/

README.md: ## convert go docs from doc.go to README.md; run with -B to force
	go get github.com/robertkrimen/godocdown/godocdown
	godocdown $(ROOTPKG) | sed "s/^--$$//" > README.md

release: req-release-type req-release-repo ## package and upload a release
	release -t $(RELEASE_TYPE) -g $(RELEASE_REPO) $(RELEASE_BRANCH) $(RELEASE_BASE)

req-release-type:
ifndef RELEASE_TYPE
	$(error RELEASE_TYPE is undefined)
endif

req-release-repo:
ifndef RELEASE_REPO
	$(error RELEASE_REPO is undefined)
endif

test: test-deps test-unit ## run unit tests
	go test -v ./lib/...

test-deps: ## install test dependencies
	go get -t -tags '$(GO_BUILD_FLAGS)' ./lib/...
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen

test-integration: export GO_BUILD_FLAGS = integration
test-integration: test-deps ## run integration tests
	go test -v -tags '$(GO_BUILD_FLAGS)' ./lib/...
