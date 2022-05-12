# NVIDIA GPU Add-on with support for GPUDirect RDMA

This branch includes changes to the NVIDIA GPU [Add-on](https://gitlab.cee.redhat.com/service/managed-tenants/-/tree/main)
that enable [GPUDirect RDMA](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/gpu-operator-rdma.html)
support just by configuring the NVIDIA GPU Add-on Custom Resource (CR).

## Motivation

The NVIDIA GPU [Add-on](https://gitlab.cee.redhat.com/service/managed-tenants/-/tree/main) aims for
seamless user experience on enabling the underlying NVIDIA hardware on [Red Hat OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift).

Currently, this add-on automates the provisioning of NVIDIA GPUs on OpenShift to enable [CUDA](https://developer.nvidia.com/cuda-toolkit).
This is accompliced by managing the deployment of the [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/overview.html)
and its CRs.

In this branch, the add-on is extended to the provisioning of NVIDIA Networking hardware, via managing
the deployment of the [NVIDIA Network Operator](https://docs.nvidia.com/networking/display/COKAN10/Network+Operator)
and its CRs, in order to enable [GPUDirect RDMA](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/gpu-operator-rdma.html).

## Goals

- prove that the NVIDIA GPU Add-on is able to handle the deployment of all the required components
- showcase all the changes required to reach support for GPUDirect RDMA
- showcase how an opinionated GPU Add-on CRD design can radically simplify the configuration of the
  required components

## Non-Goals

- create a final, production-ready GPU Add-on CRD design that includes the NVIDIA Networking
  configuration
- implement a production-ready GPU Add-on workflow for the NVIDIA Networking provisioning

## User Stories

### Story 1

As a user I would like my NVIDIA GPU and NIC with GPUDirect RDMA to be available on OpenShift, so
that I can run my GPUDirect RDMA workload without having to configure all of the components one by
one.

## Design Details

The NVIDIA GPU Add-on is now able to install the NVIDIA Network Operator via reconciling its OLM
subscription. The GPU Add-on CRD is extended with the RDMA configuration section, which includes the
configuration of the NVIDIA Network Operator [NicClusterPolicy](https://github.com/Mellanox/network-operator/blob/master/api/v1alpha1/nicclusterpolicy_types.go#L156),
[MacvlanNetwork](https://github.com/Mellanox/network-operator/blob/master/api/v1alpha1/macvlannetwork_types.go#L62).

If the deployed GPUAddon CR includes such a section, then the NVIDIA Network Operator, the NicClusterPolicy,
the MacvlanNetwork and the [ClusterPolicy](https://gitlab.com/nvidia/kubernetes/gpu-operator/-/blob/master/api/v1/clusterpolicy_types.go#L926)
CRs are configured to enable GPUDirect RDMA and deployed.

### Test Plan

For a detailed demo description check the following [doc](./hack/gpudirect-rdma-demo/README.md).

High-level description:

- setup an OpenShift 4.10 cluster with worker nodes having NVIDIA GPUs and NICs
- deploy this branch of the NVIDIA GPU Add-on on the cluster, e.g. by creating a custom CatalogSource
  containing a bundle with this version of the add-on operator. When creating the OLM Subscription
  use the [config env](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/subscription-config.md#env)
  to enable GPUDirect RDMA. Define the `GPUDIRECT_RDMA_ENABLED=true` env var. Then, the add-on will
  deploy a `GPUAddon` CR, that has GPUDirect RDMA enabled.
- this should deploy the NVIDIA Network Operator, its CRs and enable GPUDirect RDMA in the NVIDIA
  GPU Operator ClusterPolicy CR.
- run GPUDirect RDMA workload to verify the setup
