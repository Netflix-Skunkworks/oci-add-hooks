GOCMD?=go
GOBUILD?=$(GOCMD) build
GOCLEAN?=$(GOCMD) clean
GOTEST?=$(GOCMD) test
GOBIN=$(shell $(GOCMD) env GOPATH)/bin
GOBIN_TOOL?=$(shell which gobin || echo $(GOBIN)/gobin)

BINARY_NAME=oci-add-hooks
SOURCES=$(shell find . -name '*.go')

all: test build

.PHONY: build test clean

build: $(BINARY_NAME)
$(BINARY_NAME): $(SOURCES)
	$(GOBUILD) -o $(BINARY_NAME)
test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist

$(GOBIN_TOOL):
	GO111MODULE=off go get -u github.com/myitcv/gobin

.PHONY: release
release: test $(GOBIN_TOOL)
	$(GOBIN_TOOL) -run github.com/goreleaser/goreleaser@v0.106.0 release
