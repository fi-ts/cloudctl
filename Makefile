GOOS := linux
GOARCH := amd64
CGO_ENABLED := 0
BINARY := cloudctl-$(GOOS)-$(GOARCH)

SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date --rfc-3339=seconds)
VERSION := $(or ${VERSION},$(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD))

LINKMODE := $(LINKMODE) \
		 -X 'github.com/metal-stack/v.Version=$(VERSION)' \
		 -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
		 -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
		 -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'

.PHONY: build
build:
	go build \
		-tags netgo \
		-ldflags \
		"$(LINKMODE)" \
		-o bin/$(BINARY) \
		github.com/fi-ts/cloudctl

	strip bin/$(BINARY) || true
	md5sum bin/$(BINARY) > bin/$(BINARY).md5

.PHONY: test
test: build
	go test -cover ./...
