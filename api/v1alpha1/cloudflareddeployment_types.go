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
	defaultSelector    = metav1.LabelSelector{MatchLabels: defaultMatchLabels}
)

func (d *CloudflaredDeployment) ToDaemonSet(image string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: d.appMeta(),
		Spec: appsv1.DaemonSetSpec{
			Selector: &defaultSelector,
			Template: d.podTemplate(appName, image),
		},
	}
}

func (d *CloudflaredDeployment) ToDeployment(image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: d.appMeta(),
		Spec: appsv1.DeploymentSpec{
			Selector: &defaultSelector,
			Template: d.podTemplate(appName, image),
		},
	}
}

func (d *CloudflaredDeployment) appMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      d.Name,
		Namespace: d.Namespace,
	}
}

func (d *CloudflaredDeployment) podTemplate(name, image string) v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Labels: defaultMatchLabels},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  name,
				Image: image,
			}},
		},
	}
}

func (d *CloudflaredDeployment) setObjectMeta(m *metav1.ObjectMeta) {
	m.Name = d.Name
	m.Namespace = d.Namespace
}

func init() {
	SchemeBuilder.Register(&CloudflaredDeployment{}, &CloudflaredDeploymentList{})
}
