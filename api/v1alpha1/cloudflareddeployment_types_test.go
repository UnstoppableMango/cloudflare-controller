package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CloudflaredDeployment", func() {
	Context("When creating a DaemonSet", func() {
		deployment := CloudflaredDeployment{}

		Context("and kind is not set", func() {
			It("should create a DaemonSet", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
			})

			It("should use the deployment's name and namespace", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Name).To(Equal(deployment.Name))
				Expect(actual.Namespace).To(Equal(deployment.Namespace))
			})

			It("should configure the DaemonSet", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("test"))
			})
		})

		Context("and kind is DaemonSet", func() {
			BeforeEach(func() {
				By("Setting kind to DaemonSet")
				deployment.Spec.Kind = DaemonSet
			})

			It("should create a DaemonSet", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
			})

			It("should use the deployment's name and namespace", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Name).To(Equal(deployment.Name))
				Expect(actual.Namespace).To(Equal(deployment.Namespace))
			})

			It("should configure the DaemonSet", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("test"))
			})
		})

		Context("and a custom template is configured", func() {
			BeforeEach(func() {
				By("Configuring the pod template")
				deployment.Spec.Template = &v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cloudflared",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name:  "cloudflared",
							Image: "testing",
						}},
					},
				}
			})

			It("should honor the configured template", func() {
				By("Calling ToDaemonSet")
				actual := deployment.ToDaemonSet("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("testing"))
			})
		})
	})

	Context("When creating a Deployment", func() {
		deployment := CloudflaredDeployment{}

		Context("and kind is not set", func() {
			It("should create a DaemonSet", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
			})

			It("should use the deployment's name and namespace", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Name).To(Equal(deployment.Name))
				Expect(actual.Namespace).To(Equal(deployment.Namespace))
			})

			It("should configure the DaemonSet", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("test"))
			})
		})

		Context("and kind is Deployment", func() {
			BeforeEach(func() {
				By("Setting kind to Deployment")
				deployment.Spec.Kind = Deployment
			})

			It("should create a Deployment", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
			})

			It("should use the deployment's name and namespace", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Name).To(Equal(deployment.Name))
				Expect(actual.Namespace).To(Equal(deployment.Namespace))
			})

			It("should configure the Deployment", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("test"))
			})
		})

		Context("and a custom template is configured", func() {
			BeforeEach(func() {
				By("Configuring the pod template")
				deployment.Spec.Template = &v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cloudflared",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name:  "cloudflared",
							Image: "testing",
						}},
					},
				}
			})

			It("should honor the configured template", func() {
				By("Calling ToDeployment")
				actual := deployment.ToDeployment("test")

				Expect(actual).NotTo(BeNil())
				Expect(actual.Spec.Selector.MatchLabels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Labels).To(HaveKeyWithValue("app", "cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers).To(HaveLen(1))
				Expect(actual.Spec.Template.Spec.Containers[0].Name).To(Equal("cloudflared"))
				Expect(actual.Spec.Template.Spec.Containers[0].Image).To(Equal("testing"))
			})
		})
	})
})
