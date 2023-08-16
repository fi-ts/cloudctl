BINARY := cloudctl
MAINMODULE := github.com/fi-ts/cloudctl
# the builder is at https://github.com/metal-stack/builder
COMMONDIR := $(or ${COMMONDIR},../../metal-stack/builder)

-include $(COMMONDIR)/Makefile.inc

.PHONY: build-platforms
build-platforms:
	docker build --no-cache -t platforms .

.PHONY: extract-binaries
extract-binaries: build-platforms
	mkdir -p tmp
	mkdir -p result
	docker cp $(shell docker create platforms):/work/bin tmp
	mv tmp/bin/cloudctl-linux-amd64 result
	mv tmp/bin/cloudctl-windows-amd64 result
	mv tmp/bin/cloudctl-darwin-amd64 result
	mv tmp/bin/cloudctl-darwin-arm64 result
	md5sum result/cloudctl-linux-amd64 > result/cloudctl-linux-amd64.md5
	md5sum result/cloudctl-windows-amd64 > result/cloudctl-windows-amd64.md5
	md5sum result/cloudctl-darwin-amd64 > result/cloudctl-darwin-amd64.md5
	md5sum result/cloudctl-darwin-arm64 > result/cloudctl-darwin-arm64.md5
	ls -lh result
