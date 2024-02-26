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

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudflaredDeploymentKind describes the kind of deployment to create.
// +kubebuilder:validation:Enum=DaemonSet;Deployment
type CloudflaredDeploymentKind string

const (
	DaemonSet  CloudflaredDeploymentKind = "DaemonSet"
	Deployment CloudflaredDeploymentKind = "Deployment"
)

// CloudflaredDeploymentSpec defines the desired state of CloudflaredDeployment
type CloudflaredDeploymentSpec struct {
	//+kubebuilder:default:=DaemonSet
	Kind CloudflaredDeploymentKind `json:"kind,omitempty"`

	// +optional
	Template *v1.PodTemplateSpec `json:"template,omitempty"`
}

// CloudflaredDeploymentStatus defines the observed state of CloudflaredDeployment
type CloudflaredDeploymentStatus struct {
	// +optional
	State string `json:"state"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.replicas",name=Replicas,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// CloudflaredDeployment is the Schema for the cloudflareddeployments API
type CloudflaredDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudflaredDeploymentSpec   `json:"spec,omitempty"`
	Status CloudflaredDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudflaredDeploymentList contains a list of CloudflaredDeployment
type CloudflaredDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudflaredDeployment `json:"items"`
}

const appName = "cloudflared"

var (
	defaultMatchLabels = map[string]string{"app": appName}
)

func (d *CloudflaredDeployment) ToDaemonSet(image string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: d.appMeta(),
		Spec: appsv1.DaemonSetSpec{
			Selector: d.selector(),
			Template: d.podTemplate(image),
		},
	}
}

func (d *CloudflaredDeployment) ToDeployment(image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: d.appMeta(),
		Spec: appsv1.DeploymentSpec{
			Selector: d.selector(),
			Template: d.podTemplate(image),
		},
	}
}

func (d *CloudflaredDeployment) appMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: d.Name, Namespace: d.Namespace}
}

func (d *CloudflaredDeployment) selector() *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: d.matchLabels()}
}

func (d *CloudflaredDeployment) podTemplate(image string) v1.PodTemplateSpec {
	spec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: d.matchLabels(),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  appName,
				Image: image,
			}},
		},
	}

	if d.Spec.Template != nil {
		if d.Spec.Template.Labels != nil {
			spec.Labels = d.Spec.Template.Labels
		}
		if d.Spec.Template.Spec.Containers != nil {
			spec.Spec.Containers = d.Spec.Template.Spec.Containers
		}
	}

	return spec
}

func (d *CloudflaredDeployment) matchLabels() map[string]string {
	if d.Spec.Template != nil && d.Spec.Template.Labels != nil {
		// TODO: Don't blindly copy labels
		return d.Spec.Template.Labels
	}

	return defaultMatchLabels
}

func init() {
	SchemeBuilder.Register(&CloudflaredDeployment{}, &CloudflaredDeploymentList{})
}
