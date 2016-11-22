# Makefile configuration
.DEFAULT_GOAL := help

clean: ## Clears environment
	@echo $(shell date +'%H:%M:%S') "\033[0;33mRemoving old release\033[0m"
	@rm release/dogrelay* 2> /dev/null || true

test: ## Runs unit tests
	@echo $(shell date +'%H:%M:%S') "\033[0;32mRunning unit tests\033[0m"
	@go test ./...

release: clean test ## Runs all release tasks
	@GOOS="linux" GOARCH="amd64" go build -o release/dogrelay-linux64 main.go
	@GOOS="darwin" GOARCH="amd64" go build -o release/dogrelay-darwin64 main.go

help:
	@grep --extended-regexp '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
