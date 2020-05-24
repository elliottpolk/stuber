
BIN=stuber
PKG=github.com/elliottpolk/stuber
VERSION=`cat .version`
GOOS?=linux

M = $(shell printf "\033[34;1m◉\033[0m")

default: clean install ;                                              @ ## defaulting to clean and build

.PHONY: all
all: clean test install

.PHONY: build
build: ; $(info $(M) building ...)                                  @ ## build the binary
	@mkdir -p ./build/bin
	@GOOS=$(GOOS) go build \
		-ldflags "-X main.version=$(VERSION) -X main.compiled=$(date +%s)" \
		-o ./build/bin/$(BIN) \
		cmd/main.go

.PHONY: install
install: ; $(info $(M) installing locally ...)                      @ ## install the binary locally
	@GOOS=$(GOOS) go build \
		-ldflags "-X main.version=$(VERSION) -X main.compiled=$(date +%s)" \
		-o $(GOPATH)/bin/$(BIN) \
		cmd/main.go

.PHONY: test
test: ; $(info $(M) running tests ...)                   			@ ## run tests
	@go test -v -cover ./...

.PHONY: clean
clean: ; $(info $(M) running clean ...)                             @ ## clean up the old build dir
	@rm -vrf build
	@rm -vrf $(GOPATH)/bin/$(BIN)

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

