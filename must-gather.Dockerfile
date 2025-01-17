# Download the latest oc binary
FROM registry.access.redhat.com/ubi8/ubi:8.7 as builder
RUN curl -L https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable/openshift-client-linux.tar.gz | tar xz oc

# Package scripts and binaries required for must-gather
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.7

ARG VERSION
ENV VERSION=$VERSION

RUN microdnf --disableplugin=subscription-manager install -y --nodocs rsync tar && \
    microdnf --disableplugin=subscription-manager clean all

COPY --from=builder oc /usr/bin/oc
COPY must-gather/gather* /usr/bin/

ENTRYPOINT ["/bin/bash"]
