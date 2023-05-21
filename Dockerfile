# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static-debian11:nonroot

ARG BIN_FILE=./remote-work-processor

WORKDIR /
COPY ${BIN_FILE} .

USER 65532:65532

ENTRYPOINT ["/remote-work-processor"]
