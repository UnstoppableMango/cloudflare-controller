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
	"time"

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
		expectedLabels := map[string]string{"app": "cloudflared"}

		JustBeforeEach(func() {
			By("Creating the custom resource for the Kind CloudflaredDeployment")
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

			By("Reconciling the created resource")
			controllerReconciler := &CloudflaredDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
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

		It("should default to a DaemonSet", func() {
			By("Fetching the deployment")
			resource := &cfv1alpha1.CloudflaredDeployment{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			Expect(resource.Spec.Kind).To(Equal(cfv1alpha1.DaemonSet))
		})

		It("should create a DaemonSet", func() {
			By("Fetching the daemon set")
			resource := &appsv1.DaemonSet{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			Expect(resource).NotTo(BeNil())
			Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
			container := resource.Spec.Template.Spec.Containers[0]
			Expect(container.Name).To(Equal("cloudflared"))
			Expect(container.Image).To(Equal("docker.io/cloudflare/cloudflared:latest"))
		})

		It("should create a selector that matches pod labels", func() {
			By("Fetching the DaemonSet")
			resource := &appsv1.DaemonSet{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			Expect(resource).NotTo(BeNil())
			Expect(resource.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
		})

		It("should add an owner reference", func() {
			By("Fetching the DaemonSet")
			resource := &appsv1.DaemonSet{}
			Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

			controller := true
			Expect(resource).NotTo(BeNil())
			Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
				// TODO: Can any of this be pulled from somewhere else?
				APIVersion: "v1alpha1",
				Kind:       "CloudflaredDeployment",
				Name:       typeNamespacedName.Name,
				Controller: &controller,
				UID:        deployment.UID,
			}))
		})

		Context("and pod spec template is configured", func() {
			const (
				expectedImage     = "something/not/cloudflared"
				expectedContainer = "container-name"
			)

			BeforeEach(func() {
				By("Setting labels and containers")
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

			AfterEach(func() {
				By("Clearing the pod template")
				spec.Template = nil
			})

			It("should create a DaemonSet", func() {
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

			It("should create a selector that matches pod labels", func() {
				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
			})

			It("should add an owner reference", func() {
				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				controller := true
				Expect(resource).NotTo(BeNil())
				Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
					// TODO: Can any of this be pulled from somewhere else?
					APIVersion: "v1alpha1",
					Kind:       "CloudflaredDeployment",
					Name:       typeNamespacedName.Name,
					Controller: &controller,
					UID:        deployment.UID,
				}))
			})
		})

		Context("and kind is DaemonSet", func() {
			BeforeEach(func() {
				By("Setting the kind to DaemonSet")
				spec.Kind = cfv1alpha1.DaemonSet
			})

			AfterEach(func() {
				By("Clearing the kind")
				spec.Kind = ""
			})

			It("should create a DaemonSet", func() {
				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := resource.Spec.Template.Spec.Containers[0]
				Expect(container.Name).To(Equal("cloudflared"))
				Expect(container.Image).To(Equal("docker.io/cloudflare/cloudflared:latest"))
			})

			It("should create a selector that matches pod labels", func() {
				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
			})

			It("should add an owner reference", func() {
				By("Fetching the DaemonSet")
				resource := &appsv1.DaemonSet{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				controller := true
				Expect(resource).NotTo(BeNil())
				Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
					// TODO: Can any of this be pulled from somewhere else?
					APIVersion: "v1alpha1",
					Kind:       "CloudflaredDeployment",
					Name:       typeNamespacedName.Name,
					Controller: &controller,
					UID:        deployment.UID,
				}))
			})

			Context("and pod spec template is configured", func() {
				const (
					expectedImage     = "something/not/cloudflared"
					expectedContainer = "container-name"
				)

				BeforeEach(func() {
					By("Setting labels and containers")
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

				AfterEach(func() {
					By("Clearing the pod template")
					spec.Template = nil
				})

				It("should create a DaemonSet", func() {
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

				It("should create a selector that matches pod labels", func() {
					By("Fetching the DaemonSet")
					resource := &appsv1.DaemonSet{}
					Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

					Expect(resource).NotTo(BeNil())
					Expect(resource.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
				})

				It("should add an owner reference", func() {
					By("Fetching the DaemonSet")
					resource := &appsv1.DaemonSet{}
					Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

					controller := true
					Expect(resource).NotTo(BeNil())
					Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
						// TODO: Can any of this be pulled from somewhere else?
						APIVersion: "v1alpha1",
						Kind:       "CloudflaredDeployment",
						Name:       typeNamespacedName.Name,
						Controller: &controller,
						UID:        deployment.UID,
					}))
				})
			})
		})

		Context("and kind is Deployment", func() {
			BeforeEach(func() {
				By("Setting the kind to Deployment")
				spec.Kind = cfv1alpha1.Deployment
			})

			AfterEach(func() {
				By("Clearing the kind")
				spec.Kind = ""
			})

			It("should create a Deployment", func() {
				By("Fetching the deployment")
				resource := &appsv1.Deployment{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				Expect(resource).NotTo(BeNil())
				Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := resource.Spec.Template.Spec.Containers[0]
				Expect(container.Name).To(Equal("cloudflared"))
				Expect(container.Image).To(Equal("docker.io/cloudflare/cloudflared:latest"))
			})

			It("should add an owner reference", func() {
				By("Fetching the Deployment")
				resource := &appsv1.Deployment{}
				Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

				controller := true
				Expect(resource).NotTo(BeNil())
				Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
					// TODO: Can any of this be pulled from somewhere else?
					APIVersion: "v1alpha1",
					Kind:       "CloudflaredDeployment",
					Name:       typeNamespacedName.Name,
					Controller: &controller,
					UID:        deployment.UID,
				}))
			})

			Context("and pod spec template is configured", func() {
				const (
					expectedImage     = "something/not/cloudflared"
					expectedContainer = "container-name"
				)

				expectedLabels := map[string]string{"app": "cloudflared"}

				BeforeEach(func() {
					By("Setting labels and containers")
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

				AfterEach(func() {
					By("Clearing pod template")
					spec.Template = nil
				})

				It("should create a Deployment", func() {
					By("Fetching the Deployment")
					resource := &appsv1.Deployment{}
					Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

					Expect(resource).NotTo(BeNil())
					Expect(resource.Spec.Template.Labels).To(Equal(expectedLabels))
					Expect(resource.Spec.Template.Spec.Containers).To(HaveLen(1))
					container := resource.Spec.Template.Spec.Containers[0]
					Expect(container.Name).To(Equal(expectedContainer))
					Expect(container.Image).To(Equal(expectedImage))
				})

				It("should create a selector that matches pod labels", func() {
					By("Fetching the Deployment")
					resource := &appsv1.Deployment{}
					Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

					Expect(resource).NotTo(BeNil())
					Expect(resource.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
				})

				It("should add an owner reference", func() {
					By("Fetching the Deployment")
					resource := &appsv1.Deployment{}
					Eventually(k8sClient.Get(ctx, typeNamespacedName, resource)).Should(Succeed())

					controller := true
					Expect(resource).NotTo(BeNil())
					Expect(resource.OwnerReferences).To(ContainElement(metav1.OwnerReference{
						// TODO: Can any of this be pulled from somewhere else?
						APIVersion: "v1alpha1",
						Kind:       "CloudflaredDeployment",
						Name:       typeNamespacedName.Name,
						Controller: &controller,
						UID:        deployment.UID,
					}))
				})
			})
		})

		Context("and deployment is marked for deletion", func() {
			var expectedTimestamp metav1.Time

			BeforeEach(func() {
				By("Setting the deletion timestamp")
				expectedTimestamp = metav1.NewTime(time.Now())
				deployment.DeletionTimestamp = &expectedTimestamp
			})

			AfterEach(func() {
				By("Clearing the deletion timestamp")
				deployment.DeletionTimestamp = nil
			})

			It("should not leave any Deployments", func() {
				By("Attempting to fetch a matching Deployment")
				resource := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, typeNamespacedName, resource)

				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("should not leave any DaemonSets", func() {
				By("Attempting to fetch a matching DaemonSet")
				daemonSet := &appsv1.DaemonSet{}
				err := k8sClient.Get(ctx, typeNamespacedName, daemonSet)

				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			Context("and a similarly named deployment exists", func() {
				existingDeployment := appsv1.Deployment{}

				BeforeEach(func() {
					By("Setting existing deployment metadata")
					existingDeployment.Name = typeNamespacedName.Name
					existingDeployment.Namespace = typeNamespacedName.Namespace

					By("Configuring the existing deployment spec")
					existingDeployment.Spec = appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"testKey": "testValue"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"testKey": "testValue"},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{{
									Name:  "testing",
									Image: defaultImage,
								}},
							},
						},
					}
				})

				JustBeforeEach(func() {
					By("Creating the existing deployment")
					Expect(k8sClient.Create(ctx, &existingDeployment)).To(Succeed())
				})

				AfterEach(func() {
					resource := &appsv1.Deployment{}
					if err := k8sClient.Get(ctx, typeNamespacedName, resource); err == nil {
						By("Cleaning up the Deployment")
						_ = k8sClient.Delete(ctx, resource)
					}

					By("Clearing the existing deployment")
					existingDeployment = appsv1.Deployment{}
				})

				// We don't own this deployment, so we shouldn't touch it
				It("should ignore the existing deployment", func() {
					resource := &appsv1.Deployment{}
					Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
				})
			})

			Context("and an owned deployment exists", func() {
				existingDeployment := appsv1.Deployment{}

				BeforeEach(func() {
					By("Setting existing deployment metadata")
					existingDeployment.Name = typeNamespacedName.Name
					existingDeployment.Namespace = typeNamespacedName.Namespace

					By("Configuring the existing deployment spec")
					existingDeployment.Spec = appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"testKey": "testValue"},
						},
						Template: v1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"testKey": "testValue"},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{{
									Name:  "testing",
									Image: defaultImage,
								}},
							},
						},
					}

					controller := true
					By("Setting the owner reference")
					existingDeployment.OwnerReferences = append(existingDeployment.OwnerReferences, metav1.OwnerReference{
						APIVersion: "v1alpha1",
						Kind:       "CloudflaredDeployment",
						Name:       typeNamespacedName.Name,
						Controller: &controller,
						UID:        deployment.UID,
					})
				})

				JustBeforeEach(func() {
					By("Creating the existing deployment")
					Expect(k8sClient.Create(ctx, &existingDeployment)).To(Succeed())
				})

				AfterEach(func() {
					resource := &appsv1.Deployment{}
					if err := k8sClient.Get(ctx, typeNamespacedName, resource); err == nil {
						By("Cleaning up the Deployment")
						_ = k8sClient.Delete(ctx, resource)
					}

					By("Clearing the existing deployment")
					existingDeployment = appsv1.Deployment{}
				})

				It("should delete the existing deployment", func() {
					resource := &appsv1.Deployment{}
					err := k8sClient.Get(ctx, typeNamespacedName, resource)

					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})
		})
	})
})
