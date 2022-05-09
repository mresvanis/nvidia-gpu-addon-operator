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

	netopv1alpha1 "github.com/Mellanox/network-operator/api/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/scheme"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

var _ = Describe("MacvlanNetwork Resource Reconcile", Ordered, func() {
	Context("Reconcile", func() {
		common.ProcessConfig()
		rrec := &MacvlanNetworkResourceReconciler{}
		gpuAddon := addonv1alpha1.GPUAddon{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		scheme := scheme.Scheme
		Expect(netopv1alpha1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		var m netopv1alpha1.MacvlanNetwork

		Context("when RDMA is configured", func() {
			gpuAddon := &addonv1alpha1.GPUAddon{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
			}
			gpuAddon.Spec = addonv1alpha1.GPUAddonSpec{
				RDMA: &addonv1alpha1.RDMASpec{
					Devices: []addonv1alpha1.DeviceSpec{
						{
							ResourceName: "a_device_name",
							Selectors: addonv1alpha1.Selectors{
								IfNames: []string{"a_name"},
							},
						},
					},
					MacvlanNetwork: addonv1alpha1.MacvlanNetworkSpec{
						Master: "a_name",
						IPAM: addonv1alpha1.IPAMSpec{
							Range: "192.168.2.225/28",
							OmitRanges: []string{
								"192.168.2.236/30",
								"192.168.2.226/30",
							},
						},
					},
				},
			}

			It("should create the MacvlanNetwork", func() {
				c := fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects().
					Build()

				cond, err := rrec.Reconcile(context.TODO(), c, gpuAddon)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cond).To(HaveLen(1))
				Expect(cond[0].Type).To(Equal(MacvlanNetworkDeployedCondition))
				Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))

				err = c.Get(context.TODO(), types.NamespacedName{
					Name: "macvlannetwork-gpu-addon",
				}, &m)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when RDMA is not configure", func() {
			gpuAddon.Spec.RDMA = nil

			It("should not create the MacvlanNetwork", func() {
				c := fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects().
					Build()

				cond, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cond).To(HaveLen(1))
				Expect(cond[0].Type).To(Equal(MacvlanNetworkDeployedCondition))
				Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(cond[0].Reason).To(Equal("NotConfigured"))

				err = c.Get(context.TODO(), types.NamespacedName{
					Name: "macvlannetwork-gpu-addon",
				}, &m)
				Expect(err).Should(HaveOccurred())
				Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Context("Delete", func() {
		common.ProcessConfig()
		rrec := &MacvlanNetworkResourceReconciler{}

		m := &netopv1alpha1.MacvlanNetwork{
			ObjectMeta: metav1.ObjectMeta{
				Name: "macvlannetwork-gpu-addon",
			},
		}

		scheme := scheme.Scheme
		Expect(netopv1alpha1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		It("should delete the MacvlanNetwork", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(m).
				Build()

			err := rrec.Delete(context.TODO(), c)
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Get(context.TODO(), client.ObjectKey{
				Name: m.Name,
			}, m)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
