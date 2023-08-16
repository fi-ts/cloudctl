FROM ghcr.io/metal-stack/builder:latest as platforms
RUN make platforms \
 && strip bin/cloudctl-linux-amd64 bin/cloudctl

FROM alpine:3.18
LABEL maintainer="metal-stack Authors <info@metal-stack.io>"
COPY --from=platforms /work/bin/cloudctl /cloudctl
ENTRYPOINT ["/cloudctl"]
