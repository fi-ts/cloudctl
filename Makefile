BINARY := cloudctl
MAINMODULE := git.f-i-ts.de/cloud-native/cloudctl
COMMONDIR := $(or ${COMMONDIR},../common)

include $(COMMONDIR)/Makefile.inc

release:: all
