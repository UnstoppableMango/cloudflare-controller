/*
Copyright 2024 UnstoppableMango.

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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cfv1alpha1 "github.com/UnstoppableMango/cloudflare-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CloudflaredDeployment Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()
		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		deployment := &cfv1alpha1.CloudflaredDeployment{}
		spec := &cfv1alpha1.CloudflaredDeploymentSpec{}

		JustBeforeEach(func() {
			By("creating the custom resource for the Kind CloudflaredDeployment")
			err := k8sClient.Get(ctx, typeNamespacedName, deployment)
			if err != nil && errors.IsNotFound(err) {
				resource := &cfv1alpha1.CloudflaredDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: *spec,
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &cfv1alpha1.CloudflaredDeployment{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the specific resource instance of CloudflaredDeployment")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			daemonSet := &appsv1.DaemonSet{}
			if err = k8sClient.Get(ctx, typeNamespacedName, daemonSet); err == nil {
				By("Cleaning up the DaemonSet")
				_ = k8sClient.Delete(ctx, daemonSet)
			}

			deployment := &appsv1.Deployment{}
			if err = k8sClient.Get(ctx, typeNamespacedName, deployment); err == nil {
				By("Cleaning up the Deployment")
				_ = k8sClient.Delete(ctx, deployment)
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CloudflaredDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should default to a DaemonSet", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CloudflaredDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Fetching the deployment")
			resource := &cfv1alpha1.CloudflaredDeployment{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			Expect(resource.Spec.Kind).To(Equal(cfv1alpha1.DaemonSet))
		})

		It("should create a DaemonSet", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CloudflaredDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Fetching the daemon set")
			resource := &appsv1.DaemonSet{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			Expect(resource).NotTo(BeNil())
			Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
			container := resource.Spec.Template.Spec.Containers[0]
			Expect(container.Name).To(Equal("cloudflared"))
			Expect(container.Image).To(Equal("docker.io/cloudflare/cloudflared:latest"))
		})

		Context("and kind is Deployment", func() {
			BeforeEach(func() {
				By("Setting the kind to Deployment")
				spec.Kind = cfv1alpha1.Deployment
			})

			It("should create a Deployment", func() {
				By("Reconciling the created resource")
				controllerReconciler := &CloudflaredDeploymentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("Fetching the deployment")
				resource := &appsv1.Deployment{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := resource.Spec.Template.Spec.Containers[0]
				Expect(container.Name).To(Equal("cloudflared"))
				Expect(container.Image).To(Equal("docker.io/cloudflare/cloudflared:latest"))
			})
		})

		Context("and pod spec template is configured", func() {
			const (
				expectedImage     = "something/not/cloudflared"
				expectedContainer = "container-name"
			)

			expectedLabels := map[string]string{"app": "cloudflared"}

			BeforeEach(func() {
				By("Configuring the CloudflaredDeployment spec")
				spec.Kind = cfv1alpha1.DaemonSet
				spec.Template = &v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: expectedLabels},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name:  expectedContainer,
							Image: expectedImage,
						}},
					},
				}
			})

			It("should create a DaemonSet", func() {
				By("Reconciling the created resource")
				controllerReconciler := &CloudflaredDeploymentReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Template.Labels).To(Equal(expectedLabels))
				Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := resource.Spec.Template.Spec.Containers[0]
				Expect(container.Name).To(Equal(expectedContainer))
				Expect(container.Image).To(Equal(expectedImage))
			})

			// TODO: Test for selector matching custom template labels
		})
	})
})
