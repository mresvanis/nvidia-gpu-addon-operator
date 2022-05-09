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

	netopv1alpha1 "github.com/Mellanox/network-operator/api/v1alpha1"
	wheretypes "github.com/k8snetworkplumbingwg/whereabouts/pkg/types"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

const (
	MacvlanNetworkDeployedCondition = "MacvlanNetworkDeployed"

	macvlanNetworkName = "macvlannetwork-gpu-addon"
)

type MacvlanNetworkResourceReconciler struct{}

var _ ResourceReconciler = &MacvlanNetworkResourceReconciler{}

func (r *MacvlanNetworkResourceReconciler) Reconcile(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) ([]metav1.Condition, error) {

	logger := log.FromContext(ctx, "Reconcile Step", "MacvlanNetwork CR")
	conditions := []metav1.Condition{}

	if gpuAddon.Spec.RDMA == nil {
		conditions = append(conditions, r.getDeployedConditionNotConfigured())

		logger.Info("MacvlanNetwork CR will not be reconciled as GPUAddon RDMA is not configured")

		return conditions, nil
	}

	existingMVN := &netopv1alpha1.MacvlanNetwork{}

	err := c.Get(ctx, client.ObjectKey{
		Name: macvlanNetworkName,
	}, existingMVN)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		conditions = append(conditions, r.getDeployedConditionFetchFailed())
		return conditions, err
	}

	m := &netopv1alpha1.MacvlanNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name: macvlanNetworkName,
		},
	}

	if exists {
		m = existingMVN
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), c, m, func() error {
		return r.setDesiredMacvlanNetwork(c, m, gpuAddon)
	})

	if err != nil {
		conditions = append(conditions, r.getDeployedConditionCreateFailed())
		return conditions, err
	}

	conditions = append(conditions, r.getDeployedConditionCreateSuccess())

	logger.Info("MacvlanNetwork reconciled successfully", "name", m.Name, "result", res)

	return conditions, nil
}

func (r *MacvlanNetworkResourceReconciler) setDesiredMacvlanNetwork(
	c client.Client,
	m *netopv1alpha1.MacvlanNetwork,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if m == nil {
		return errors.New("macvnetwork cannot be nil")
	}

	m.Spec = netopv1alpha1.MacvlanNetworkSpec{}

	m.Spec.Mtu = 1500
	m.Spec.Mode = "bridge"
	m.Spec.Master = "ens2f0np0"

	ipamConfig, err := getMacvlanNetworkIPAMConfig(gpuAddon.Spec.RDMA)
	if err != nil {
		return err
	}

	m.Spec.IPAM = ipamConfig

	return nil
}

func (r *MacvlanNetworkResourceReconciler) Delete(ctx context.Context, c client.Client) error {
	m := &netopv1alpha1.MacvlanNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name: macvlanNetworkName,
		},
	}

	err := c.Delete(ctx, m)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete MacvlanNetwork %s: %w", m.Name, err)
	}

	return nil
}

func (r *MacvlanNetworkResourceReconciler) getDeployedConditionFetchFailed() metav1.Condition {
	return common.NewCondition(
		MacvlanNetworkDeployedCondition,
		metav1.ConditionTrue,
		"FetchCrFailed",
		"Failed to fetch MacvlanNetwork CR")
}

func (r *MacvlanNetworkResourceReconciler) getDeployedConditionCreateFailed() metav1.Condition {
	return common.NewCondition(
		MacvlanNetworkDeployedCondition,
		metav1.ConditionTrue,
		"CreateCrFailed",
		"Failed to create MacvlanNetwork CR")
}

func (r *MacvlanNetworkResourceReconciler) getDeployedConditionCreateSuccess() metav1.Condition {
	return common.NewCondition(
		MacvlanNetworkDeployedCondition,
		metav1.ConditionTrue,
		"CreateCrSuccess",
		"MacvlanNetwork CR deployed successfully")
}

func (r *MacvlanNetworkResourceReconciler) getDeployedConditionNotConfigured() metav1.Condition {
	return common.NewCondition(
		MacvlanNetworkDeployedCondition,
		metav1.ConditionTrue,
		"NotConfigured",
		"GPUAddon RDMA is not configured, the MacvlanNetwork CR won't be deployed")
}

func getMacvlanNetworkIPAMConfig(rdmaSpec *addonv1alpha1.RDMASpec) (string, error) {
	config := &wheretypes.IPAMConfig{}

	config.Type = "whereabouts"
	config.Range = rdmaSpec.MacvlanNetwork.IPAM.Range
	config.OmitRanges = rdmaSpec.MacvlanNetwork.IPAM.OmitRanges

	raw, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
