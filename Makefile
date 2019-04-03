NAME := nomad-driver-singularity
PKG := github.com/sylabs/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=
# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/dist

GOOSARCHES=linux/amd64

GO=go

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package.

$(NAME): $(wildcard *.go) $(wildcard */*.go)
	@echo "+ $@"
	$(V)GO111MODULE=on GOOS=linux $(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) ./cmd/driver/main.go

.PHONY: static
static: ## Builds a static executable.
	@echo "+ $@"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
				-tags "$(BUILDTAGS) static_build" \
				${GO_LDFLAGS_STATIC} -o $(NAME) .

all: clean build fmt lint test vet

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes.
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests.
	@echo "+ $@"
	@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes.
	@echo "+ $@"
	@$(GO) vet -printfuncs Error,ErrorDepth,Errorf,Errorln,Exit,ExitDepth,Exitf,Exitln,Fatal,FatalDepth,Fatalf,Fatalln,Info,InfoDepth,Infof,Infoln,Warning,WarningDepth,Warningf,Warningln -all ./...

.PHONY: cover
cover: ## Runs go test with coverage.
	@echo "" > coverage.txt
	@for d in $(shell $(GO) list ./... | grep -v vendor); do \
		$(GO) test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
endef

.PHONY: clean
clean: ## Cleanup any build binaries or packages.
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) -r $(BUILDDIR)

dep:
	$(V)GO111MODULE=on go mod download
	$(V)GO111MODULE=on go mod vendor
	$(V)GO111MODULE=on go mod tidy
