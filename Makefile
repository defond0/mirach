buil.PHONY: clean clean-test clean-pyc clean-build docs help
.DEFAULT_GOAL := help
define BROWSER_PYSCRIPT
import os, webbrowser, sys
try:
	from urllib import pathname2url
except:
	from urllib.request import pathname2url

webbrowser.open("file://" + pathname2url(os.path.abspath(sys.argv[1])))
endef
export BROWSER_PYSCRIPT

define PRINT_HELP_PYSCRIPT
import re, sys

for line in sys.stdin:
	match = re.match(r'^([\w_-]+):.*?## (.*)$$', line)
	if match:
		target, help = match.groups()
		print("%-20s %s" % (target, help))
endef
export PRINT_HELP_PYSCRIPT
BROWSER := python -c "$$BROWSER_PYSCRIPT"

# project variables
PROJECT_NAME := mirach

help:
	@python -c "$$PRINT_HELP_PYSCRIPT" < $(MAKEFILE_LIST)

build: test build-linux build-windows ## build all os and arch

build-linux: build-linux-386 build-linux-amd64 ## build all linux arch

build-linux-386: ## build linux 386
	GOOS=linux GOARCH=386 go build -o mirach_linux_386

build-linux-amd64: ## build linux amd64
	GOOS=linux GOARCH=amd64 go build -o mirach_linux_amd64

build-windows: build-windows-386 build-windows-amd64 ## build all windows arch

build-windows-386: ## build windows 386
	GOOS=windows GOARCH=386 go build -o mirach_win_386.exe

build-windows-amd64: ## build windows amd64
	GOOS=windows GOARCH=amd64 go build -o mirach_win_amd64.exe

clean: clean-build ## clean all

clean-build: ## remove build artifacts
	rm -rf mirach_*

deploy-docs: docs ## deploy docs to S3 bucket
	aws s3 sync ./docs/_build/html/ s3://***REMOVED***/$(PROJECT_NAME)/

docs: ## generate docs
	echo "Figure out docs"

lint: ## gofmt goimports
	gofmt *.go
	-goimport *.go

test: test-integration ## run tests
	@echo "Tests Run"

test-integration: ## run integration tests
	go build -race .
	go test -v -tags=integration .

test-race-condition: ## run and observe race
	go build -race .
	./mirach -vv
