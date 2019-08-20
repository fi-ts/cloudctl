BINARY := cloudctl
MAINMODULE := git.f-i-ts.de/cloud-native/cloudctl
COMMONDIR := $(or ${COMMONDIR},../common)
SWAGGER_VERSION := $(or ${SWAGGER_VERSION},v0.19.0)

include $(COMMONDIR)/Makefile.inc

release:: all

