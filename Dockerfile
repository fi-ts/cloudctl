FROM ghcr.io/metal-stack/builder:latest
RUN make platforms \
 && strip bin/cloudctl-linux-amd64
