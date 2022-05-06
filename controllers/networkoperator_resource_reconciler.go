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
	"errors"
	"fmt"

	netopconsts "github.com/Mellanox/network-operator/pkg/consts"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

const (
	NetworkOperatorDeployedCondition = "NetworkOperatorDeployed"

	networkOperatorPackageName      = "nvidia-network-operator"
	networkOperatorSubscriptionName = "nvidia-network-operator"
)

type NetworkOperatorResourceReconciler struct{}

var _ ResourceReconciler = &NetworkOperatorResourceReconciler{}

func (r *NetworkOperatorResourceReconciler) Reconcile(
	ctx context.Context,
	client client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) ([]metav1.Condition, error) {

	conditions := []metav1.Condition{}

	if err := r.reconcileSubscription(ctx, client, gpuAddon); err != nil {
		conditions = append(conditions, r.getDeployedConditionFailed(err))
		return conditions, err
	}

	if err := r.reconcileNamespace(ctx, client, gpuAddon); err != nil {
		conditions = append(conditions, r.getDeployedConditionFailed(err))
		return conditions, err
	}

	conditions = append(conditions, r.getDeployedConditionSuccess())

	return conditions, nil
}

func (r *NetworkOperatorResourceReconciler) reconcileSubscription(
	ctx context.Context,
	client client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	logger := log.FromContext(ctx, "Reconcile Step", "NetworkOperator Subscription")
	existingSubscription := &operatorsv1alpha1.Subscription{}

	err := client.Get(ctx, types.NamespacedName{
		Namespace: gpuAddon.Namespace,
		Name:      networkOperatorSubscriptionName,
	}, existingSubscription)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	s := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gpuAddon.Namespace,
			Name:      networkOperatorSubscriptionName,
		},
	}

	if exists {
		s = existingSubscription

		// if s.Status.InstalledCSV != "" {
		// 	NetworkOperatorSubscriptionInstalled.WithLabelValues(s.Status.CurrentCSV, s.Status.InstalledCSV).Set(1)
		// } else {
		// 	NetworkOperatorSubscriptionInstalled.WithLabelValues("", "").Set(0)
		// }
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), client, s, func() error {
		return r.setDesiredSubscription(client, s, gpuAddon)
	})

	if err != nil {
		return err
	}

	logger.Info("NetworkOperator Subscription reconciled successfully",
		"name", s.Name,
		"namespace", s.Namespace,
		"result", res)

	return nil
}

func (r *NetworkOperatorResourceReconciler) reconcileNamespace(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	logger := log.FromContext(ctx, "Reconcile Step", "NetworkOperator Resources Namespace")
	existingNS := &corev1.Namespace{}

	err := c.Get(ctx, client.ObjectKey{
		Name: netopconsts.NetworkOperatorResourceNamespace,
	}, existingNS)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: netopconsts.NetworkOperatorResourceNamespace,
		},
	}

	if exists {
		ns = existingNS
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), c, ns, func() error {
		return r.setDesiredNamespace(c, ns, gpuAddon)
	})

	if err != nil {
		return err
	}

	logger.Info("NetworkOperator Resources Namespace reconciled successfully",
		"name", ns.Name,
		"result", res)

	return nil
}

func (r *NetworkOperatorResourceReconciler) setDesiredSubscription(
	client client.Client,
	s *operatorsv1alpha1.Subscription,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if s == nil {
		return errors.New("subscription cannot be nil")
	}

	s.Spec = &operatorsv1alpha1.SubscriptionSpec{
		CatalogSource:          "certified-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Channel:                "v1.1.0",
		Package:                networkOperatorPackageName,
		InstallPlanApproval:    operatorsv1alpha1.ApprovalAutomatic,
	}

	if err := ctrl.SetControllerReference(gpuAddon, s, client.Scheme()); err != nil {
		return err
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) setDesiredNamespace(
	client client.Client,
	ns *corev1.Namespace,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if ns == nil {
		return errors.New("namespace cannot be nil")
	}

	ns.ObjectMeta = metav1.ObjectMeta{
		Name: netopconsts.NetworkOperatorResourceNamespace,
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) Delete(ctx context.Context, c client.Client) error {
	err := r.deleteSubscription(ctx, c)
	if err != nil {
		return err
	}

	err = r.deleteCSV(ctx, c)
	if err != nil {
		return err
	}

	err = r.deleteNamespace(ctx, c)
	if err != nil {
		return err
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) deleteSubscription(ctx context.Context, c client.Client) error {
	s := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.GlobalConfig.AddonNamespace,
			Name:      networkOperatorSubscriptionName,
		},
	}

	err := c.Delete(ctx, s)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete NetworkOperator Subscription %s: %w", s.Name, err)
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) deleteNamespace(ctx context.Context, c client.Client) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: netopconsts.NetworkOperatorResourceNamespace,
		},
	}

	err := c.Delete(ctx, ns)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete NetworkOperator Resources Namespace %s: %w", ns.Name, err)
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) deleteCSV(ctx context.Context, c client.Client) error {
	csv, err := common.GetCsvWithPrefix(c, common.GlobalConfig.AddonNamespace, networkOperatorPackageName)
	if err != nil {
		return err
	}

	err = c.Delete(ctx, csv)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete NetworkOperator CSV %s: %w", csv.Name, err)
	}

	return nil
}

func (r *NetworkOperatorResourceReconciler) getDeployedConditionFailed(err error) metav1.Condition {
	return common.NewCondition(
		NetworkOperatorDeployedCondition,
		metav1.ConditionTrue,
		"Failed",
		err.Error())
}

func (r *NetworkOperatorResourceReconciler) getDeployedConditionSuccess() metav1.Condition {
	return common.NewCondition(
		NetworkOperatorDeployedCondition,
		metav1.ConditionTrue,
		"Success",
		"NetworkOperator deployed successfully")
}
