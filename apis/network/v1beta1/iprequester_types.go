/*
Copyright 2023.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IPRequesterSpec defines the desired state of IPRequester
type IPRequesterSpec struct {
	// +kubebuilder:validation:Optional
	HostName string `json:"hostname,omitempty"`
}

// Network Annotation schema
type Network struct {
	// +kubebuilder:validation:Required
	// Network Name
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// Subnet Name
	SubnetName string `json:"subnetName"`

	// +kubebuilder:validation:Optional
	// Fixed Ip
	FixedIP string `json:"fixedIP,omitempty"`
}

// IPRequesterStatus defines the observed state of IPRequester
type IPRequesterStatus struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Allocated bool `json:"allocated"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IPRequester is the Schema for the iprequesters API
type IPRequester struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPRequesterSpec   `json:"spec,omitempty"`
	Status IPRequesterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPRequesterList contains a list of IPRequester
type IPRequesterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPRequester `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPRequester{}, &IPRequesterList{})
}
