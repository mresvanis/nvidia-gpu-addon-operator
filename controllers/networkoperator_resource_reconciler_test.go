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

	netopconsts "github.com/Mellanox/network-operator/pkg/consts"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/scheme"
	corev1 "k8s.io/api/core/v1"
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

var _ = Describe("NetworkOperator Resource Reconcile", Ordered, func() {
	Context("Reconcile", func() {
		common.ProcessConfig()
		rrec := &NetworkOperatorResourceReconciler{}
		gpuAddon := addonv1alpha1.GPUAddon{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
		}

		scheme := scheme.Scheme
		Expect(operatorsv1alpha1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		var s operatorsv1alpha1.Subscription
		var ns corev1.Namespace

		It("should create the NetworOperator Subscription", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects().
				Build()

			_, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Get(context.TODO(), types.NamespacedName{
				Namespace: gpuAddon.Namespace,
				Name:      "nvidia-network-operator",
			}, &s)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should create the NetworOperator Resource Namespace", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects().
				Build()

			_, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Get(context.TODO(), client.ObjectKey{
				Name: netopconsts.NetworkOperatorResourceNamespace,
			}, &ns)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Delete", func() {
		common.ProcessConfig()
		rrec := &NetworkOperatorResourceReconciler{}

		s := &operatorsv1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nvidia-network-operator",
				Namespace: common.GlobalConfig.AddonNamespace,
			},
		}
		csv := &operatorsv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nvidia-network-operator.v1.1.0",
				Namespace: common.GlobalConfig.AddonNamespace,
			},
		}
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: netopconsts.NetworkOperatorResourceNamespace,
			},
		}

		scheme := scheme.Scheme
		Expect(operatorsv1alpha1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		It("should delete the NetworkOperator Subscription, CSV and Resource Namespace", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(s, csv, ns).
				Build()

			err := rrec.Delete(context.TODO(), c)
			Expect(err).ShouldNot(HaveOccurred())

			err = c.Get(context.TODO(), types.NamespacedName{
				Name:      s.Name,
				Namespace: s.Namespace,
			}, s)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())

			err = c.Get(context.TODO(), types.NamespacedName{
				Name:      csv.Name,
				Namespace: csv.Namespace,
			}, csv)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())

			err = c.Get(context.TODO(), client.ObjectKey{
				Name: netopconsts.NetworkOperatorResourceNamespace,
			}, csv)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
