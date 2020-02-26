# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: compile
compile: fmt vet pkg/liiklus/LiiklusService.pb.go ## Compile target binaries
	go build .

pkg/liiklus/LiiklusService.pb.go: LiiklusService.proto
	protoc -I . LiiklusService.proto --go_out=plugins=grpc:pkg/liiklus

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -timeout 30s -coverprofile cover.out

# Run go fmt against code
.PHONY: fmt
fmt: goimports
	$(GOIMPORTS) -w --local github.com/projectriff pkg/

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# find or download goimports, download goimports if necessary
goimports:
ifeq (, $(shell which goimports))
	cp go.mod go.mod~ && cp go.sum go.sum~ # avoid go.* mutations from go get
	go get golang.org/x/tools/cmd/goimports@release-branch.go1.13
	mv go.mod~ go.mod && mv go.sum~ go.sum
GOIMPORTS=$(GOBIN)/goimports
else
GOIMPORTS=$(shell which goimports)
endif

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
