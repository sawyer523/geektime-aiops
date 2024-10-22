/*
Copyright 2024.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicaitonSpec defines the desired state of Applicaiton
type ApplicaitonSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Deployment ApplicationDeployment    `json:"deployment"`
	Service    corev1.ServiceSpec       `json:"service"`
	Ingress    networkingv1.IngressSpec `json:"ingress"`
	ConfigMap  *corev1.ConfigMap        `json:"configMap"`
}

type ApplicationDeployment struct {
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
	Port     int32  `json:"port"`
}

// ApplicaitonStatus defines the observed state of Applicaiton
type ApplicaitonStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Applicaiton is the Schema for the applicaitons API
type Applicaiton struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicaitonSpec   `json:"spec,omitempty"`
	Status ApplicaitonStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicaitonList contains a list of Applicaiton
type ApplicaitonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Applicaiton `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Applicaiton{}, &ApplicaitonList{})
}
