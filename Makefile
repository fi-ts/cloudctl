BINARY := cloudctl
MAINMODULE := github.com/fi-ts/cloudctl
COMMONDIR := $(or ${COMMONDIR},../common)

include $(COMMONDIR)/Makefile.inc

release:: all
