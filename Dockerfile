FROM golang:1.20-bullseye as builder

WORKDIR /autopi

# Copy the Go Modules manifests
COPY . .

# Build
RUN CGO_ENABLED=0 make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static-debian11:nonroot

WORKDIR /
COPY --from=builder /autopi/remote-work-processor .

USER 65532:65532

ENTRYPOINT ["/remote-work-processor"]
