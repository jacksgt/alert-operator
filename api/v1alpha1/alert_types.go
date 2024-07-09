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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertSpec defines the desired state of Alert
type AlertSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

// AlertStatus defines the observed state of Alert
type AlertStatus struct {
	// State describes if the alert is currently active or not.
	State string `json:"state,omitempty"`
	// Annotations contains key-value data associated to the alert.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels contains key-value data associated to the alert.
	Labels map[string]string `json:"labels,omitempty"`
	// Describes since which timestamp the alert is active.
	Since string `json:"since,omitempty"` // TODO: use a proper timestamp
	// The current value of alert expression.
	Value string	`json:"value,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Alert is the Schema for the alerts API
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Since",type=string,JSONPath=`.status.since`
// https://book.kubebuilder.io/reference/generating-crd.html#additional-printer-columns
type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertSpec   `json:"spec,omitempty"`
	Status AlertStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertList contains a list of Alert
type AlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alert{}, &AlertList{})
}
