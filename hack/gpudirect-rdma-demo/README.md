# NVIDIA GPU Add-on Operator GPUDirect RDMA Demo

## Requirements

- OpenShift 4.10
- Worker nodes with NVIDIA GPU and NIC devices

## Versions used

This branch of the NVIDIA GPU Add-on Operator uses the following component versions:

- NVIDIA GPU Operator `v1.10.0`
  - NVIDIA driver image `nvcr.io/nvidia/driver:510.47.03-rhcos4.9`
- NVIDIA Network Operator `1.2.0`
  - MLNX_OFED driver version image `nvcr.io/nvidia/mellanox/mofed-5.6-1.0.3.3:rhcos4.10-amd64`
  - Kubernetes RDMA Shared Device Plugin image `nvcr.io/nvidia/cloud-native/k8s-rdma-shared-dev-plugin:v1.3.2`
  - SR-IOV Network Device Plugin image `ghcr.io/k8snetworkplumbingwg/sriov-network-device-plugin:v3.4.0`

## Install the NVIDIA GPU Add-on with GPUDirect RDMA Support

```shell
$ oc apply -f catalogsource.yaml
catalogsource.operators.coreos.com/nvidia-addon-catalog configured

# wait until the additional catalog is up and running
$ oc get pods -n openshift-marketplace -l olm.catalogSource=nvidia-addon-catalog
NAME                         READY   STATUS    RESTARTS   AGE
nvidia-addon-catalog-dqgrg   1/1     Running   0          20s

# install the NVIDIA GPU Add-on operator Subscription
$ oc apply -f addon.yaml
namespace/redhat-nvidia-gpu-addon created
operatorgroup.operators.coreos.com/redhat-nvidia-gpu-addon created
subscription.operators.coreos.com/nvidia-gpu-addon-operator created
```

## Check NFD, NVIDIA GPU Operator and NVIDIA Network Operator Resources

```shell
# check NVIDIA Network Operator Resources
$ oc get -n nvidia-network-operator-resources pods
NAME                        READY   STATUS    RESTARTS   AGE
mofed-rhcos4.10-ds-j5gkc    1/1     Running   0          6m42s
mofed-rhcos4.10-ds-m6jz8    1/1     Running   0          6m42s
rdma-shared-dp-ds-8g692     1/1     Running   0          11s
rdma-shared-dp-ds-pn5jq     1/1     Running   0          31s
sriov-device-plugin-2bt4j   1/1     Running   0          31s
sriov-device-plugin-ffvf2   1/1     Running   0          11s

# check NFD, NVIDIA GPU Add-on Operator, NVIDIA GPU Operator, NVIDIA Network Operator
$ oc get -n redhat-nvidia-gpu-addon pods
NAME                                                          READY   STATUS      RESTARTS      AGE
controller-manager-78d777f4db-zjh4h                           2/2     Running     0             24m
gpu-feature-discovery-9rbbm                                   1/1     Running     0             9m
gpu-feature-discovery-wjjcp                                   1/1     Running     0             9m
gpu-operator-6f88b8f54f-np6nl                                 1/1     Running     0             9m9s
nfd-controller-manager-69c597fdbf-zswvs                       2/2     Running     1 (24m ago)   24m
nfd-master-fd7pr                                              1/1     Running     0             24m
nfd-master-szfxf                                              1/1     Running     0             24m
nfd-master-xr4sw                                              1/1     Running     0             24m
nfd-worker-64dzm                                              1/1     Running     0             18m
nfd-worker-rwwlw                                              1/1     Running     0             18m
nvidia-container-toolkit-daemonset-gzdr6                      1/1     Running     0             9m1s
nvidia-container-toolkit-daemonset-pd2p4                      1/1     Running     0             9m1s
nvidia-cuda-validator-7n9qq                                   0/1     Completed   0             41s
nvidia-cuda-validator-slpr2                                   0/1     Completed   0             56s
nvidia-dcgm-exporter-sln2w                                    1/1     Running     0             9m
nvidia-dcgm-exporter-zq9wm                                    1/1     Running     0             9m
nvidia-dcgm-tfm7x                                             1/1     Running     0             9m1s
nvidia-dcgm-wnxh5                                             1/1     Running     0             9m1s
nvidia-device-plugin-daemonset-827b7                          1/1     Running     0             9m1s
nvidia-device-plugin-daemonset-rzfg9                          1/1     Running     0             9m1s
nvidia-device-plugin-validator-kdxwt                          0/1     Completed   0             45s
nvidia-device-plugin-validator-rgjcg                          0/1     Completed   0             30s
nvidia-driver-daemonset-49.84.202204072350-0-v8wpp            3/3     Running     0             9m1s
nvidia-driver-daemonset-49.84.202204072350-0-xw7sn            3/3     Running     0             9m1s
nvidia-network-operator-controller-manager-5dc9bd9bd7-p4s4d   2/2     Running     0             9m11s
nvidia-node-status-exporter-clgj7                             1/1     Running     0             9m2s
nvidia-node-status-exporter-qqx6c                             1/1     Running     0             9m2s
nvidia-operator-validator-99nm5                               1/1     Running     0             9m1s
nvidia-operator-validator-vsgqp                               1/1     Running     0             9m1s
```

## Run GPUDirect RDMA Workload

```shell
# run the server pod
$ oc apply -f rdma-gpudirect-server.yaml
pod/rdma-gpudirect-workload-server created

# wait for the server pod to be up and running
$ oc get pod rdma-gpudirect-workload-server
NAME                             READY   STATUS    RESTARTS   AGE
rdma-gpudirect-workload-server   1/1     Running   0          87s

# run the client pod
$ oc apply -f rdma-gpudirect-client.yaml
pod/rdma-gpudirect-workload-client created

# check the client logs
$ oc logs -f rdma-gpudirect-workload-client
...
---------------------------------------------------------------------------------------
initializing CUDA
Listing all CUDA devices in system:
CUDA device 0: PCIe address is 41:00

Picking device No. 0
[pid = 1, dev = 0] device name = [Tesla T4]
creating CUDA Ctx
making it the current CUDA Ctx
cuMemAlloc() of a 33554432 bytes GPU buffer
allocated GPU buffer address at 00007f71e8000000 pointer=0x7f71e8000000
---------------------------------------------------------------------------------------
                    RDMA_Write BW Test
 Dual-port       : OFF          Device         : mlx5_0
 Number of qps   : 2            Transport type : IB
 Connection type : RC           Using SRQ      : OFF
 PCIe relax order: Unsupported
 ibv_wr* API     : OFF
 TX depth        : 128
 CQ Moderation   : 100
 Mtu             : 1024[B]
 Link type       : Ethernet
 GID index       : 4
 Max inline data : 0[B]
 rdma_cm QPs     : ON
 Data ex. method : rdma_cm
---------------------------------------------------------------------------------------
 local address: LID 0000 QPN 0x00e4 PSN 0xbaa67d
 GID: 00:00:00:00:00:00:00:00:00:00:255:255:192:168:02:226
 local address: LID 0000 QPN 0x00e5 PSN 0x109b23
 GID: 00:00:00:00:00:00:00:00:00:00:255:255:192:168:02:226
 remote address: LID 0000 QPN 0x00e4 PSN 0xbaa67d
 GID: 00:00:00:00:00:00:00:00:00:00:255:255:192:168:02:225
 remote address: LID 0000 QPN 0x00e5 PSN 0x109b23
 GID: 00:00:00:00:00:00:00:00:00:00:255:255:192:168:02:225
---------------------------------------------------------------------------------------
 #bytes     #iterations    BW peak[Gb/sec]    BW average[Gb/sec]   MsgRate[Mpps]
 2          10000           0.044399            0.044359            2.772442
 4          10000           0.089271            0.089067            2.783335
 8          10000            0.18               0.18               2.827731
 16         10000            0.36               0.35               2.757577
 32         10000            0.71               0.71               2.767234
 64         10000            1.41               1.41               2.758215
 128        10000            2.86               2.85               2.786095
 256        10000            4.49               4.45               2.171864
 512        10000            8.75               8.69               2.122719
 1024       10000            17.34              17.33              2.115637
 2048       10000            22.89              22.38              1.366241
 4096       10000            22.83              22.76              0.694454
 8192       10000            22.95              22.95              0.350152
 16384      10000            22.98              22.97              0.175263
 32768      10000            23.08              23.00              0.087720
 65536      10000            23.09              23.03              0.043933
 131072     10000            23.29              23.22              0.022147
 262144     10000            23.08              23.03              0.010983
 524288     10000            23.07              23.03              0.005490
 1048576    10000            23.03              23.02              0.002744
 2097152    10000            23.03              23.02              0.001372
 4194304    10000            23.02              23.00              0.000686
 8388608    10000            23.01              22.99              0.000343
---------------------------------------------------------------------------------------
deallocating RX GPU buffer 00007f71e8000000
destroying current CUDA Ctx
```

## Cleanup
```shell
$ oc delete gpuaddons.nvidia.addons.rh-ecosystem-edge.io -n redhat-nvidia-gpu-addon nvidia-gpu-addon
gpuaddon.nvidia.addons.rh-ecosystem-edge.io "nvidia-gpu-addon" delete

# wait for the NFD operator to completely remove the NFD instance
$ oc logs -f -n redhat-nvidia-gpu-addon -l control-plane=controller-manager -c manager
...
I0517 13:14:30.219193       1 nodefeaturediscovery_controller.go:51] [ClusterRoleBinding nfd-master resource has been deleted.]
I0517 13:14:35.219400       1 nodefeaturediscovery_controller.go:51] [Worker ServiceAccount resource has been deleted.]
I0517 13:14:40.220116       1 nodefeaturediscovery_controller.go:51] [Master ServiceAccount resource has been deleted.]
I0517 13:14:45.220583       1 nodefeaturediscovery_controller.go:51] [SecurityContextConstraints nfd-worker resource has been deleted.]
I0517 13:14:50.221012       1 nodefeaturediscovery_controller.go:51] [nfd-worker config map resource has been deleted.]
I0517 13:14:50.221054       1 nodefeaturediscovery_controller.go:51] [Deletion appears to have succeeded, but running a secondary check to ensure resources are cleaned up]
I0517 13:14:50.221138       1 nodefeaturediscovery_controller.go:51] [Secondary check passed.  Removing finalizer if it exists.]
I0517 13:14:50.237050       1 nodefeaturediscovery_controller.go:51] [Finalizer was found and successfully removed.]
I0517 13:14:50.237165       1 nodefeaturediscovery_controller.go:51] [Fetch the NodeFeatureDiscovery instance]
I0517 13:14:50.237214       1 nodefeaturediscovery_controller.go:55] resource has been deleted%!(EXTRA []interface {}=[req ocp-gpu-addon got ])

# delete the GPU Add-on namespace
$ oc delete ns redhat-nvidia-gpu-addon
namespace "redhat-nvidia-gpu-addon" deleted
```
