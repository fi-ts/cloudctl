BINARY := cloudctl
MAINMODULE := github.com/mreiger/cloudctl
COMMONDIR := $(or ${COMMONDIR},../common)

include $(COMMONDIR)/Makefile.inc

release:: all
