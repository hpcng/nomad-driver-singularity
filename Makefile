NAME := nomad-driver-singularity
PKG := github.com/sylabs/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=
# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

GOOSARCHES=linux/amd64

GO=$(shell which go)

all: clean build fmt lint test vet

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package.

$(NAME): $(wildcard *.go) $(wildcard */*.go)
	@echo "+ $@"
	$(V)GO111MODULE=on GOOS=linux $(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) ./cmd/driver/main.go

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@echo "+ $@"
	@gofmt -s -l . | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes.
	@echo "+ $@"
	@golint ./... | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests.
	@echo "+ $@"
	$(V)GO111MODULE=on $(GO) test -v -tags "$(BUILDTAGS) cgo" ./...

.PHONY: vet
vet: ## Verifies `go vet` passes.
	@echo "+ $@"
	@$(GO) vet -printfuncs Error,ErrorDepth,Errorf,Errorln,Exit,ExitDepth,Exitf,Exitln,Fatal,FatalDepth,Fatalf,Fatalln,Info,InfoDepth,Infof,Infoln,Warning,WarningDepth,Warningf,Warningln -all ./...

.PHONY: cover
cover: ## Runs go test with coverage.
	@echo "" > coverage.txt
	$(V)GO111MODULE=on $(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...; \

.PHONY: clean
clean: ## Cleanup any build binaries or packages.
	@echo "+ $@"
	$(RM) $(NAME)

dep:
	$(V)GO111MODULE=on go mod download
