/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertificateSpec defines the desired state of Certificate.
type CertificateSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf", message="dnsName is immutable"
	DNSName string `json:"dnsName"` // dnsName is immutable, since you cant change the domain of a certificate without regenerating it.

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`\d+d`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf", message="validity is immutable"
	Validity string `json:"validity"` // validity is also immutable, since you cant change the validity period without regenerating

	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`
}

const (
	StatusReady  = "Ready"
	StatusFailed = "Failed"
)

// CertificateStatus defines the observed state of Certificate.
type CertificateStatus struct {
	Status string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Certificate is the Schema for the certificates API.
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateList contains a list of Certificate.
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
