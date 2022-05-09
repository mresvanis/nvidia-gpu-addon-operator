/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	k8srdmatypes "github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	netopv1alpha1 "github.com/Mellanox/network-operator/api/v1alpha1"
	netopconsts "github.com/Mellanox/network-operator/pkg/consts"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

const (
	NicClusterPolicyDeployedCondition = "NicClusterPolicyDeployed"
)

type NicClusterPolicyResourceReconciler struct{}

var _ ResourceReconciler = &NicClusterPolicyResourceReconciler{}

func (r *NicClusterPolicyResourceReconciler) Reconcile(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) ([]metav1.Condition, error) {

	logger := log.FromContext(ctx, "Reconcile Step", "NicClusterPolicy CR")
	conditions := []metav1.Condition{}

	if gpuAddon.Spec.RDMA == nil {
		conditions = append(conditions, r.getDeployedConditionNotConfigured())
		logger.Info("NicClusterPolicy CR will not be reconciled as GPUAddon RDMA is not configured")
		return conditions, nil
	}

	existingCP := &netopv1alpha1.NicClusterPolicy{}

	err := c.Get(ctx, client.ObjectKey{
		Name: netopconsts.NicClusterPolicyResourceName,
	}, existingCP)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		conditions = append(conditions, r.getDeployedConditionFetchFailed())
		return conditions, err
	}

	cp := &netopv1alpha1.NicClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: netopconsts.NicClusterPolicyResourceName,
		},
	}

	if exists {
		cp = existingCP
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), c, cp, func() error {
		return r.setDesiredNicClusterPolicy(c, cp, gpuAddon)
	})

	if err != nil {
		conditions = append(conditions, r.getDeployedConditionCreateFailed())
		return conditions, err
	}

	conditions = append(conditions, r.getDeployedConditionCreateSuccess())

	logger.Info("NicClusterPolicy reconciled successfully", "name", cp.Name, "result", res)

	return conditions, nil
}

func (r *NicClusterPolicyResourceReconciler) setDesiredNicClusterPolicy(
	c client.Client,
	cp *netopv1alpha1.NicClusterPolicy,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if cp == nil {
		return errors.New("nicclusterpolicy cannot be nil")
	}

	cp.Spec = netopv1alpha1.NicClusterPolicySpec{}

	cp.Spec.OFEDDriver = &netopv1alpha1.OFEDDriverSpec{
		ImageSpec: netopv1alpha1.ImageSpec{
			Image:      "mofed",
			Repository: "nvcr.io/nvidia/mellanox",
			Version:    "5.6-1.0.3.3",
		},
	}

	rdmaDevicePluginConfig, err := getK8sRdmaSharedDevicePluginConfig(gpuAddon.Spec.RDMA)
	if err != nil {
		return err
	}

	cp.Spec.RdmaSharedDevicePlugin = &netopv1alpha1.DevicePluginSpec{
		ImageSpec: netopv1alpha1.ImageSpec{
			Image:      "k8s-rdma-shared-dev-plugin",
			Repository: "nvcr.io/nvidia/cloud-native",
			Version:    "v1.3.2",
		},
		Config: rdmaDevicePluginConfig,
	}

	cp.Spec.SriovDevicePlugin = &netopv1alpha1.DevicePluginSpec{
		ImageSpec: netopv1alpha1.ImageSpec{
			Image:      "sriov-network-device-plugin",
			Repository: "ghcr.io/k8snetworkplumbingwg",
			Version:    "v3.4.0",
		},
		Config: `
		{
			"resourceList": [
				{
					"resourcePrefix": "nvidia.com",
					"resourceName": "hostdev",
					"selectors": {
						"vendors": ["15b3"],
						"isRdma": true
					}
				}
			]
		}`,
	}

	return nil
}

func (r *NicClusterPolicyResourceReconciler) Delete(ctx context.Context, c client.Client) error {
	cp := &netopv1alpha1.NicClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: netopconsts.NicClusterPolicyResourceName,
		},
	}

	err := c.Delete(ctx, cp)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete NicClusterPolicy %s: %w", cp.Name, err)
	}

	return nil
}

func (r *NicClusterPolicyResourceReconciler) getDeployedConditionFetchFailed() metav1.Condition {
	return common.NewCondition(
		NicClusterPolicyDeployedCondition,
		metav1.ConditionTrue,
		"FetchCrFailed",
		"Failed to fetch NicClusterPolicy CR")
}

func (r *NicClusterPolicyResourceReconciler) getDeployedConditionCreateFailed() metav1.Condition {
	return common.NewCondition(
		NicClusterPolicyDeployedCondition,
		metav1.ConditionTrue,
		"CreateCrFailed",
		"Failed to create NicClusterPolicy CR")
}

func (r *NicClusterPolicyResourceReconciler) getDeployedConditionCreateSuccess() metav1.Condition {
	return common.NewCondition(
		NicClusterPolicyDeployedCondition,
		metav1.ConditionTrue,
		"CreateCrSuccess",
		"NicClusterPolicy deployed successfully")
}

func (r *NicClusterPolicyResourceReconciler) getDeployedConditionNotConfigured() metav1.Condition {
	return common.NewCondition(
		NicClusterPolicyDeployedCondition,
		metav1.ConditionTrue,
		"NotConfigured",
		"GPUAddon RDMA is not configured, the NicClusterPolicy CR won't be deployed")
}

func getK8sRdmaSharedDevicePluginConfig(rdmaSpec *addonv1alpha1.RDMASpec) (string, error) {
	config := &k8srdmatypes.UserConfigList{}

	configs := []k8srdmatypes.UserConfig{}
	for _, device := range rdmaSpec.Devices {
		config := k8srdmatypes.UserConfig{}
		config.ResourceName = device.ResourceName
		config.Selectors = k8srdmatypes.Selectors{
			IfNames: device.Selectors.IfNames,
		}
		config.RdmaHcaMax = 1000
		configs = append(configs, config)
	}

	config.ConfigList = configs

	raw, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
