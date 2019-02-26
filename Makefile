# Makefile configuration
.DEFAULT_GOAL := help
.PHONY: clean test deps release

clean: ## Clears environment
	@echo $(shell date +'%H:%M:%S') "\033[0;33mRemoving old release\033[0m"
	@mkdir -p release
	@rm -rf release/*

test: ## Runs unit tests
	@echo $(shell date +'%H:%M:%S') "\033[0;32mRunning unit tests\033[0m"
	@go test ./...

deps: ## Download required dependencies
	@echo $(shell date +'%H:%M:%S') "\033[0;32mDownloading dependencies\033[0m"
	@go get github.com/spf13/cobra
	@go get github.com/mono83/xray
	@go get github.com/mono83/validate
	@go get github.com/mono83/udpwriter
	@go get github.com/stretchr/testify/assert

release: clean deps test ## Runs all release tasks
	@echo $(shell date +'%H:%M:%S') "\033[0;32mCompiling Linux version\033[0m"
	@GOOS="linux" GOARCH="amd64" go build -o release/dogrelay-linux64 main.go
	@echo $(shell date +'%H:%M:%S') "\033[0;32mCompiling MacOS version\033[0m"
	@GOOS="darwin" GOARCH="amd64" go build -o release/dogrelay-darwin64 main.go

help:
	@grep --extended-regexp '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
