package gpuaddon

import (
	"context"
	"errors"
	"fmt"

	gpuv1 "github.com/NVIDIA/gpu-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

const (
	ClusterPolicyDeployedCondition = "ClusterPolicyDeployed"

	dcgmMetricsConfigMapName = "custom-dcgm-metrics"

	defaultDcgmMetrics = `
DCGM_FI_PROF_GR_ENGINE_ACTIVE, gauge, gpu utilization.
DCGM_FI_DEV_MEM_COPY_UTIL, gauge, mem utilization.
DCGM_FI_DEV_ENC_UTIL, gauge, enc utilization.
DCGM_FI_DEV_DEC_UTIL, gauge, dec utilization.
DCGM_FI_DEV_POWER_USAGE, gauge, power usage.
DCGM_FI_DEV_POWER_MGMT_LIMIT_MAX, gauge, power mgmt limit.
DCGM_FI_DEV_GPU_TEMP, gauge, gpu temp.
DCGM_FI_DEV_SM_CLOCK, gauge, sm clock.
DCGM_FI_DEV_MAX_SM_CLOCK, gauge, max sm clock.
DCGM_FI_DEV_MEM_CLOCK, gauge, mem clock.
DCGM_FI_DEV_MAX_MEM_CLOCK, gauge, max mem clock.
`
)

type ClusterPolicyResourceReconciler struct{}

var _ ResourceReconciler = &ClusterPolicyResourceReconciler{}

func (r *ClusterPolicyResourceReconciler) Reconcile(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) ([]metav1.Condition, error) {

	logger := log.FromContext(ctx, "Reconcile Step", "ClusterPolicy")
	conditions := []metav1.Condition{}

	if err := r.reconcileDcgmExporterConfigMap(ctx, c, gpuAddon); err != nil {
		conditions = append(conditions, r.getDeployedConditionFailed(err))
		return conditions, err
	}

	if err := r.reconcileClusterPolicy(ctx, c, gpuAddon); err != nil {
		conditions = append(conditions, r.getDeployedConditionFailed(err))
		return conditions, err
	}

	conditions = append(conditions, r.getDeployedConditionSuccess())

	logger.Info("ClusterPolicy reconciled successfully",
		"name", common.GlobalConfig.ClusterPolicyName)

	return conditions, nil
}

func (r *ClusterPolicyResourceReconciler) reconcileClusterPolicy(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	logger := log.FromContext(ctx, "Reconcile Step", "ClusterPolicy CR")
	existingCP := &gpuv1.ClusterPolicy{}

	err := c.Get(ctx, client.ObjectKey{
		Name: common.GlobalConfig.NfdCrName,
	}, existingCP)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	cp := &gpuv1.ClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GlobalConfig.ClusterPolicyName,
		},
	}

	if exists {
		cp = existingCP
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), c, cp, func() error {
		return r.setDesiredClusterPolicy(c, cp, gpuAddon)
	})

	if err != nil {
		return err
	}

	logger.Info("ClusterPolicy reconciled successfully",
		"name", cp.Name,
		"result", res)

	return nil
}

func (r *ClusterPolicyResourceReconciler) reconcileDcgmExporterConfigMap(
	ctx context.Context,
	c client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	logger := log.FromContext(ctx, "Reconcile Step", "DCGM Exporter ConfigMap")
	existingCM := &corev1.ConfigMap{}

	err := c.Get(ctx, client.ObjectKey{
		Name:      dcgmMetricsConfigMapName,
		Namespace: common.GlobalConfig.AddonNamespace,
	}, existingCM)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dcgmMetricsConfigMapName,
			Namespace: common.GlobalConfig.AddonNamespace,
		},
	}

	if exists {
		cm = existingCM
	}

	res, err := controllerutil.CreateOrPatch(ctx, c, cm, func() error {
		return r.setDesiredDcgmExporterConfigMap(c, cm, gpuAddon)
	})

	if err != nil {
		return err
	}

	logger.Info("ConfigMap reconciled successfully",
		"name", cm.Name,
		"result", res)

	return nil
}

func (r *ClusterPolicyResourceReconciler) setDesiredDcgmExporterConfigMap(
	c client.Client,
	cm *corev1.ConfigMap,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if cm == nil {
		return errors.New("configmap cannot be nil")
	}

	cm.Data = map[string]string{
		"dcgm-metrics.csv": defaultDcgmMetrics,
	}

	if err := ctrl.SetControllerReference(gpuAddon, cm, c.Scheme()); err != nil {
		return err
	}

	return nil
}

func (r *ClusterPolicyResourceReconciler) setDesiredClusterPolicy(
	c client.Client,
	cp *gpuv1.ClusterPolicy,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if cp == nil {
		return errors.New("clusterpolicy cannot be nil")
	}

	enabled := true
	disabled := false

	cp.Spec = gpuv1.ClusterPolicySpec{}

	cp.Spec.Operator = gpuv1.OperatorSpec{
		DefaultRuntime: gpuv1.CRIO,
	}

	cp.Spec.PSP = gpuv1.PSPSpec{
		Enabled: &disabled,
	}

	cp.Spec.Toolkit = gpuv1.ToolkitSpec{
		Enabled: &enabled,
	}

	cp.Spec.DCGM = gpuv1.DCGMSpec{
		Enabled: &enabled,
	}

	cp.Spec.DCGMExporter = gpuv1.DCGMExporterSpec{
		MetricsConfig: &gpuv1.DCGMExporterMetricsConfig{
			Name: dcgmMetricsConfigMapName,
		},
	}

	cp.Spec.MIGManager = gpuv1.MIGManagerSpec{
		Enabled: &enabled,
	}

	cp.Spec.NodeStatusExporter = gpuv1.NodeStatusExporterSpec{
		Enabled: &enabled,
	}

	cp.Spec.MIG = gpuv1.MIGSpec{
		Strategy: gpuv1.MIGStrategySingle,
	}

	cp.Spec.Validator = gpuv1.ValidatorSpec{
		Env: []corev1.EnvVar{
			{Name: "WITH_WORKLOAD", Value: "true"},
		},
	}

	cp.Spec.Driver = gpuv1.DriverSpec{
		Enabled: &enabled,
	}

	cp.Spec.Driver.UseOpenShiftDriverToolkit = &enabled

	cp.Spec.Driver.GPUDirectRDMA = &gpuv1.GPUDirectRDMASpec{
		Enabled: &disabled,
	}

	cp.Spec.Driver.LicensingConfig = &gpuv1.DriverLicensingConfigSpec{
		NLSEnabled: &disabled,
	}

	// IMPORTANT: cannot set a namespaced owner as a reference on a cluster-scoped resource.
	// "cluster-scoped resource must not have a namespace-scoped owner, owner's namespace x"
	// if err := ctrl.SetControllerReference(gpuAddon, cp, c.Scheme()); err != nil {
	// 	return err
	// }

	return nil
}

func (r *ClusterPolicyResourceReconciler) Delete(ctx context.Context, c client.Client) error {
	cp := &gpuv1.ClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: common.GlobalConfig.ClusterPolicyName,
		},
	}

	err := c.Delete(ctx, cp)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ClusterPolicy %s: %w", cp.Name, err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dcgmMetricsConfigMapName,
			Namespace: common.GlobalConfig.AddonNamespace,
		},
	}

	err = c.Delete(ctx, cm)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ConfigMap %s: %w", cm.Name, err)
	}

	return nil
}

func (r *ClusterPolicyResourceReconciler) getDeployedConditionFailed(err error) metav1.Condition {
	return common.NewCondition(
		ClusterPolicyDeployedCondition,
		metav1.ConditionFalse,
		"Failed",
		err.Error())
}

func (r *ClusterPolicyResourceReconciler) getDeployedConditionSuccess() metav1.Condition {
	return common.NewCondition(
		ClusterPolicyDeployedCondition,
		metav1.ConditionTrue,
		"Success",
		"ClusterPolicy deployed successfully")
}
