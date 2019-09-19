BINARY := cloudctl
MAINMODULE := git.f-i-ts.de/cloud-native/cloudctl
COMMONDIR := $(or ${COMMONDIR},../common)
SWAGGER_VERSION := $(or ${SWAGGER_VERSION},v0.19.0)

include $(COMMONDIR)/Makefile.inc

release:: generate-client all

.PHONY: generate-client
generate-client:
	rm -rf api
	mkdir -p api
	GO111MODULE=off swagger generate client -f cloud-api.json -t api --skip-validation