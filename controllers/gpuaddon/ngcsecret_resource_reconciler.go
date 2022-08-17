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

package gpuaddon

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

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
	NGCSecretDeployedCondition = "NGCSecretDeployed"

	secretName = "ngc-secret"

	addonParametersSecretName = "addon-nvidia-gpu-addon-parameters"

	// https://gitlab.com/nvidia/kubernetes/gpu-operator/-/blob/master/scripts/install-gpu-operator-nvaie.sh#L44
	dockerServer   = "nvcr.io"
	dockerUsername = "$oauthtoken"

	// https://gitlab.cee.redhat.com/service/managed-tenants/-/blob/main/docs/ocm/addon_parameters.md#accessing-parameter-values
	ngcAPIKeyParamName = "ngc-api-key"
	ngcEmailParamName  = "ngc-email"
)

type NGCSecretResourceReconciler struct{}

var _ ResourceReconciler = &NGCSecretResourceReconciler{}

func (r *NGCSecretResourceReconciler) Reconcile(
	ctx context.Context,
	client client.Client,
	gpuAddon *addonv1alpha1.GPUAddon) ([]metav1.Condition, error) {

	logger := log.FromContext(ctx, "Reconcile Step", "NGC Secret")
	conditions := []metav1.Condition{}

	addonParametersSecret := &corev1.Secret{}

	err := client.Get(ctx, types.NamespacedName{
		Namespace: gpuAddon.Namespace,
		Name:      addonParametersSecretName,
	}, addonParametersSecret)

	exists := !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		conditions = append(conditions, r.getDeployedConditionFetchParametersFailed())
		return conditions, err
	}

	if !exists {
		conditions = append(conditions, r.getDeployedConditionParametersNotPresent())

		logger.Info("NGC secret will not be reconciled as the add-on parameters secret is not present",
			"NGC secret name", secretName,
			"add-on parameters secret", addonParametersSecretName,
			"namespace", gpuAddon.Namespace)

		return conditions, nil
	}

	if isValid, err := r.isAddonParametersSecretValid(addonParametersSecret); !isValid {
		conditions = append(conditions, r.getDeployedConditionParametersNotValid())

		logger.Info("NGC secret will not be reconciled as the add-on parameters secret is invalid",
			"NGC secret name", secretName,
			"add-on parameters secret", addonParametersSecretName,
			"namespace", gpuAddon.Namespace,
			"error", err)

		return conditions, nil
	}

	existingSecret := &corev1.Secret{}

	err = client.Get(ctx, types.NamespacedName{
		Namespace: gpuAddon.Namespace,
		Name:      secretName,
	}, existingSecret)

	exists = !k8serrors.IsNotFound(err)
	if err != nil && !k8serrors.IsNotFound(err) {
		conditions = append(conditions, r.getDeployedConditionFetchFailed())
		return conditions, err
	}

	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gpuAddon.Namespace,
			Name:      secretName,
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	if exists {
		s = existingSecret
	}

	res, err := controllerutil.CreateOrPatch(context.TODO(), client, s, func() error {
		return r.setDesiredNGCSecret(client, s, addonParametersSecret, gpuAddon)
	})

	if err != nil {
		conditions = append(conditions, r.getDeployedConditionCreateFailed())
		return conditions, err
	}

	conditions = append(conditions, r.getDeployedConditionCreateSuccess())

	logger.Info("NGC secret reconciled successfully",
		"name", s.Name,
		"namespace", s.Namespace,
		"result", res)

	return conditions, nil
}

func (r *NGCSecretResourceReconciler) setDesiredNGCSecret(
	client client.Client,
	s *corev1.Secret,
	addonParametersSecret *corev1.Secret,
	gpuAddon *addonv1alpha1.GPUAddon) error {

	if s == nil {
		return errors.New("secret cannot be nil")
	}

	if addonParametersSecret == nil {
		return errors.New("addon parameters secret cannot be nil")
	}

	password := string(addonParametersSecret.Data[ngcAPIKeyParamName])
	email := string(addonParametersSecret.Data[ngcEmailParamName])
	auth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dockerUsername, password)))

	dockerconfig := map[string]map[string]map[string]string{
		"auths": map[string]map[string]string{
			dockerServer: map[string]string{
				"username": dockerUsername,
				"password": password,
				"email":    email,
				"auth":     auth,
			},
		},
	}
	dockerconfigjson, err := json.Marshal(dockerconfig)
	if err != nil {
		return err
	}

	s.ObjectMeta = metav1.ObjectMeta{
		Name:      s.Name,
		Namespace: s.Namespace,
	}
	s.StringData = map[string]string{".dockerconfigjson": string(dockerconfigjson)}

	return ctrl.SetControllerReference(gpuAddon, s, client.Scheme())
}

func (r *NGCSecretResourceReconciler) Delete(ctx context.Context, c client.Client) (bool, error) {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.GlobalConfig.AddonNamespace,
			Name:      secretName,
		},
	}

	if err := c.Delete(ctx, s); err != nil {
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to delete secret %s: %w", s.Name, err)
	}

	return true, nil
}

func (r *NGCSecretResourceReconciler) getDeployedConditionParametersNotPresent() metav1.Condition {
	return common.NewCondition(
		NGCSecretDeployedCondition,
		metav1.ConditionTrue,
		"AddonParametersSecretNotPresent",
		"Add-on parameters secret is not present")
}

func (r *NGCSecretResourceReconciler) getDeployedConditionParametersNotValid() metav1.Condition {
	return common.NewCondition(
		NGCSecretDeployedCondition,
		metav1.ConditionTrue,
		"AddonParametersSecretNotValid",
		"Add-on parameters secret contains invalid parameters")
}

func (r *NGCSecretResourceReconciler) getDeployedConditionFetchParametersFailed() metav1.Condition {
	return common.NewCondition(
		NGCSecretDeployedCondition,
		metav1.ConditionTrue,
		"FetchAddonParametersFailed",
		"Failed to fetch add-on parameters Secret")
}

func (r *NGCSecretResourceReconciler) getDeployedConditionFetchFailed() metav1.Condition {
	return common.NewCondition(
		NGCSecretDeployedCondition,
		metav1.ConditionTrue,
		"FetchSecretFailed",
		"Failed to fetch NGC Secret")
}

func (r *NGCSecretResourceReconciler) getDeployedConditionCreateFailed() metav1.Condition {
	return common.NewCondition(
		SubscriptionDeployedCondition,
		metav1.ConditionTrue,
		"CreateSecretFailed",
		"Failed to create NGC Secret")
}

func (r *NGCSecretResourceReconciler) getDeployedConditionCreateSuccess() metav1.Condition {
	return common.NewCondition(
		NGCSecretDeployedCondition,
		metav1.ConditionTrue,
		"CreateSecretSuccess",
		"NGC Secret deployed successfully")
}

func (r *NGCSecretResourceReconciler) isAddonParametersSecretValid(aps *corev1.Secret) (bool, error) {
	ngcEmail := aps.Data[ngcEmailParamName]
	if len(ngcEmail) <= 0 {
		return false, fmt.Errorf("%s parameter is not present", ngcEmailParamName)
	}

	ngcAPIKey := aps.Data[ngcAPIKeyParamName]
	if len(ngcAPIKey) <= 0 {
		return false, fmt.Errorf("%s parameter is not present", ngcAPIKeyParamName)
	}

	return true, nil
}
