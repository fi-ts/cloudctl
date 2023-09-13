GOOS := linux
GOARCH := amd64
CGO_ENABLED := 1
TAGS := -tags 'netgo'
BINARY := cloudctl-$(GOOS)-$(GOARCH)

SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date --iso-8601=seconds) # this format is parsable with Go RFC3339
VERSION := $(or ${VERSION},$(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD))

ifeq ($(CGO_ENABLED),1)
ifeq ($(GOOS),linux)
	LINKMODE := -linkmode external -extldflags '-static -s -w'
	TAGS := -tags 'osusergo netgo static_build'
endif
endif

LINKMODE := $(LINKMODE) \
		 -X 'github.com/metal-stack/v.Version=$(VERSION)' \
		 -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
		 -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
		 -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'

.PHONY: build
build:
	go build \
		$(TAGS) \
		-ldflags \
		"$(LINKMODE)" \
		-o bin/$(BINARY) \
		github.com/fi-ts/cloudctl

	md5sum bin/$(BINARY) > bin/$(BINARY).md5

.PHONY: test
test: build
	go test -cover ./...
