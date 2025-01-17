# Build the manager binary
From registry.access.redhat.com/ubi8/go-toolset:1.18 as builder
WORKDIR /opt/app-root/src

# Copy the Makefile, Go Modules manifests and vendored dependencies
COPY --chown=1001:0 Makefile Makefile
COPY --chown=1001:0 go.mod go.mod
COPY --chown=1001:0 go.sum go.sum
COPY --chown=1001:0 vendor vendor

# Copy the go source
COPY --chown=1001:0 main.go main.go
COPY --chown=1001:0 api/ api/
COPY --chown=1001:0 controllers/ controllers/
COPY --chown=1001:0 internal/ internal/

# Build
RUN make build

# Use UBI8 Micro as minimal base image to package the manager binary
# Refer to https://www.redhat.com/en/blog/introduction-ubi-micro for more details
FROM registry.access.redhat.com/ubi8/ubi-micro:8.7
COPY --from=builder /opt/app-root/src/bin/manager /usr/local/bin/manager
USER 1001

ENTRYPOINT ["/usr/local/bin/manager"]
