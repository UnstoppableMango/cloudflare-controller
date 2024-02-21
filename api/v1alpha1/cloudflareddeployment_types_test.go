package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		})
	})
})
