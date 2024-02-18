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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:printcolumn:JSONPath=".status.replicas",name=Replicas,type=int
//+kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// CloudflaredDeploymentSpec defines the desired state of CloudflaredDeployment
type CloudflaredDeploymentSpec struct {
	//+kubebuilder:validation:Enum=Deployment;DaemonSet
	//+kubebuilder:default:=DaemonSet
	Kind string `json:"kind,omitempty"`

	// +optional
	Template *v1.PodTemplateSpec `json:"template,omitempty"`
}

// CloudflaredDeploymentStatus defines the observed state of CloudflaredDeployment
type CloudflaredDeploymentStatus struct {
	// +optional
	Replicas int `json:"replicas"`

	// +optional
	State string `json:"state"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

func init() {
	SchemeBuilder.Register(&CloudflaredDeployment{}, &CloudflaredDeploymentList{})
}
