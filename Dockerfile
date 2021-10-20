FROM metalstack/builder:latest as builder
RUN make platforms \
 && strip bin/cloudctl-linux-amd64 bin/cloudctl

FROM alpine:3.14
LABEL maintainer="metal-stack Authors <info@metal-stack.io>"
COPY --from=builder /work/bin/cloudctl /cloudctl
ENTRYPOINT ["/cloudctl"]
