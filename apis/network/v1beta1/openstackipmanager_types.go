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
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OpenStackIPManagerSpec defines the desired state of OpenStackIPManager
type OpenStackIPManagerSpec struct {
	// +kubebuilder:validation:Optional
	Dummy string `json:"dummy,omitempty"`
}

// OpenStackIPManagerStatus defines the observed state of OpenStackIPManager
type OpenStackIPManagerStatus struct {
	// Conditions Status conditions
	Conditions condition.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

// OpenStackIPManager is the Schema for the openstackipmanagers API
type OpenStackIPManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenStackIPManagerSpec   `json:"spec,omitempty"`
	Status OpenStackIPManagerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenStackIPManagerList contains a list of OpenStackIPManager
type OpenStackIPManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenStackIPManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenStackIPManager{}, &OpenStackIPManagerList{})
}
