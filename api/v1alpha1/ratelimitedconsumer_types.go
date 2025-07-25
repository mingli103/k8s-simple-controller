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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RateLimitedConsumerSpec defines the desired state of RateLimitedConsumer.
type RateLimitedConsumerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	RateLimit   RateLimit   `json:"rateLimit"`
	TargetRoute TargetRoute `json:"targetRoute"`
}

// RateLimit defines rate limiting config
// type RateLimit struct {
// 	Name string `json:"name"`
// }

type RateLimit struct {
	Names []string `json:"names"`
}

// TargetRoute references the route to annotate
type TargetRoute struct {
	Name string `json:"name"`
	// Optional: add Namespace string `json:"namespace,omitempty"` if needed
}

// RateLimitedConsumerStatus defines the observed state of RateLimitedConsumer.
type RateLimitedConsumerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
	ObservedRoute string             `json:"observedRoute,omitempty"`
	PluginApplied string             `json:"pluginApplied,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RateLimitedConsumer is the Schema for the ratelimitedconsumers API.
type RateLimitedConsumer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitedConsumerSpec   `json:"spec,omitempty"`
	Status RateLimitedConsumerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RateLimitedConsumerList contains a list of RateLimitedConsumer.
type RateLimitedConsumerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimitedConsumer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimitedConsumer{}, &RateLimitedConsumerList{})
}
