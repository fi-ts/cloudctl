BINARY := cloudctl
MAINMODULE := github.com/fi-ts/cloudctl
# the builder is at https://github.com/metal-stack/builder
COMMONDIR := $(or ${COMMONDIR},../../metal-stack/builder)

include $(COMMONDIR)/Makefile.inc

.PHONY: all
all:: markdown

release:: all

.PHONY: markdown
markdown:
	rm -rf docs
	mkdir -p docs
	bin/cloudctl markdown
