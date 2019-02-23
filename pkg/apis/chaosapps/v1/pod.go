/*
Copyright 2019 The Kubernetes Authors.

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

//go:generate go run ../../../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generate.deepcopy -i . -h ../../../../hack/boilerplate.go.txt
package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "chaosapps.metamagical.io", Version: "v1"}

func AddToScheme(scheme *runtime.Scheme) {
	scheme.AddKnownTypes(SchemeGroupVersion, &ChaosPod{}, &ChaosPodList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
}

// ChaosPodSpec defines the desired state of ChaosPod
type ChaosPodSpec struct {
	Template corev1.PodTemplateSpec `json:"template"`
	// +optional
	NextStop metav1.Time `json:"nextStop,omitempty"`
}

// ChaosPodStatus defines the observed state of ChaosPod.
// It should always be reconstructable from the state of the cluster and/or outside world.
type ChaosPodStatus struct {
	LastRun metav1.Time `json:"lastRun,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChaosPod is the Schema for the randomjobs API
// +kubebuilder:printcolumn:name="next stop",type="string",JSONPath=".spec.nextStop",format="date"
// +kubebuilder:printcolumn:name="last run",type="string",JSONPath=".status.lastRun",format="date"
// +k8s:openapi-gen=true
type ChaosPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosPodSpec   `json:"spec,omitempty"`
	Status ChaosPodStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChaosPodList contains a list of ChaosPod
type ChaosPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosPod `json:"items"`
}
