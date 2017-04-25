PHONY: all archives clean clean-build deploy-docs docs help lint test
.DEFAULT_GOAL := help

# project variables
PROJECT_NAME := mirach
VERSION := $(shell git describe --always --dirty)

# helper variables
BUILDDIR := ./_build
ARCDIR := $(BUILDDIR)/arc
BINDIR := $(BUILDDIR)/bin
DOWNLOADLOC := s3://***REMOVED***/mirach
DOWNLOADSNAPLOC := $(DOWNLOADLOC)/SNAPSHOT
DOWNLOADSRELEASELOC := $(DOWNLOADLOC)/RELEASE
ROOTPKG := gitlab.eng.cleardata.com/dash/mirach
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

archives: $(ARC_TARGETS) ## archive all builds

clean: clean-build clean-mocks ## clean all

clean-build: ## remove build artifacts
	rm -rf $(BUILDDIR)

clean-mocks: ## remove mock artifacts
	rm -rf $(GOPATH)/src/cleardata.com/mirach/.mocks

deploy-docs: docs ## deploy docs to S3 bucket
	aws s3 sync ./docs/html s3://***REMOVED***/$(PROJECT_NAME)/

docs: docs/html/index.html ## generate docs

docs/html/index.html:
	mkdir -p docs/html/
	godoc -url "http://localhost:6060/pkg/$(ROOTPKG)" | \
	sed 's/\(href="\)\(\/lib\)/\1https:\/\/golang.org\2/g' > docs/html/index.html


install: install-build-deps ## install to GOPATH
	go install -v -ldflags=$(LDFLAGS)

install-build-deps: ## install go dependencies
	go get ./...

install-test-deps: ## install go dependencies
	go get -t -tags '$(GO_BUILD_FLAGS)' ./...

lint: ## gofmt goimports
	gofmt *.go
	-goimport *.go

mqtt-paho-mocks:
	mkdir .mocks
	mockery -inpkg -dir $(GOPATH)/src/github.com/eclipse/paho.mqtt.golang/  -all  -output $(GOPATH)/src/cleardata.com/mirach/.mocks/

publish: ## publish all current build archives
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADLOC)"
	aws s3 sync $(ARCDIR)/ $(DOWNLOADLOC)/

publish-snap: ## publish current build archives to snap shot location
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADSNAPLOC)"
	aws s3 sync $(ARCDIR)/ $(DOWNLOADSNAPLOC)/

publish-release: publish ## publish current build archives to snap shot location
	@echo "syncing contents of $(ARCDIR) to $(DOWNLOADSRELEASELOC)"
	aws s3 sync $(ARCDIR)/ $(DOWNLOADSRELEASELOC)/

README.md: ## convert go docs from doc.go to README.md; run with -B to force
	go get github.com/robertkrimen/godocdown/godocdown
	godocdown $(ROOTPKG) | sed "s/^--$$//" > README.md

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

test: test-unit ## run unit tests

test-all: test-unit test-integration

test-integration: export GO_BUILD_FLAGS = integration
test-integration: install-test-deps ## run integration tests
	go test -v -tags '$(GO_BUILD_FLAGS)' ./...

test-unit: export GO_BUILD_FLAGS = unit
test-unit: install-test-deps ## run unit tests
	go test -v -tags '$(GO_BUILD_FLAGS)' ./...
