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

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

var _ = Describe("NGC Secret Resource Reconcile", Ordered, func() {
	Context("Reconcile", func() {
		common.ProcessConfig()
		rrec := &NGCSecretResourceReconciler{}
		gpuAddon := addonv1alpha1.GPUAddon{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: common.GlobalConfig.AddonNamespace,
			},
		}

		scheme := scheme.Scheme

		var s corev1.Secret

		Context("when the addonParametersSecret is present", func() {
			Context("and contains valid parameters", func() {
				addonParametersSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "addon-nvidia-gpu-addon-parameters",
						Namespace: common.GlobalConfig.AddonNamespace,
					},
					Data: map[string][]byte{
						"ngc-api-key": []byte("a-key"),
						"ngc-email":   []byte("an-email"),
					},
				}

				It("should create the NGC Secret", func() {
					c := fake.
						NewClientBuilder().
						WithScheme(scheme).
						WithRuntimeObjects(addonParametersSecret).
						Build()

					_, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
					Expect(err).ShouldNot(HaveOccurred())

					err = c.Get(context.TODO(), types.NamespacedName{
						Namespace: gpuAddon.Namespace,
						Name:      secretName,
					}, &s)
					Expect(err).ShouldNot(HaveOccurred())
				})
			})

			Context("and contains invalid parameters", func() {
				invalidAddonParametersSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "addon-nvidia-gpu-addon-parameters",
						Namespace: common.GlobalConfig.AddonNamespace,
					},
					Data: map[string][]byte{
						"ngc-email": []byte("an-email"),
					},
				}

				It("should not create the NGC Secret", func() {
					c := fake.
						NewClientBuilder().
						WithScheme(scheme).
						WithRuntimeObjects(invalidAddonParametersSecret).
						Build()

					_, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
					Expect(err).ShouldNot(HaveOccurred())

					err = c.Get(context.TODO(), types.NamespacedName{
						Namespace: gpuAddon.Namespace,
						Name:      secretName,
					}, &s)
					Expect(err).Should(HaveOccurred())
				})
			})
		})

		Context("when the addonParametersSecret is not present", func() {
			It("should not create the NGC Secret", func() {
				c := fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects().
					Build()

				_, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
				Expect(err).ShouldNot(HaveOccurred())

				err = c.Get(context.TODO(), types.NamespacedName{
					Namespace: gpuAddon.Namespace,
					Name:      secretName,
				}, &s)
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Context("Delete", func() {
		common.ProcessConfig()
		rrec := &NGCSecretResourceReconciler{}

		s := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: common.GlobalConfig.AddonNamespace,
				Name:      secretName,
			},
		}

		scheme := scheme.Scheme

		It("should delete the NGC Secret", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(s).
				Build()

			deleted, err := rrec.Delete(context.TODO(), c)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(deleted).To(BeTrue())

			err = c.Get(context.TODO(), types.NamespacedName{
				Namespace: s.Namespace,
				Name:      s.Name,
			}, s)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
