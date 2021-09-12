/*
Copyright 2021.

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

const (
	PhasePending = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseDone    = "DONE"
	PhaseFinish  = "FINISH"
)

type Schedule struct {
	// Start time for scheduling
	Start string `json:"start,omitempty"`
	// End time for scheduling
	End string `json:"end,omitempty"`
}

// ScheduledScalerSpec defines the desired state of ScheduledScaler
type ScheduledScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Schedule defines the time between scaling up and down
	Schedule Schedule `json:"schedule,omitempty"`
	// DeploymentName defines target of deployment
	DeploymentName string `json:"deploymentName,omitempty"`
	// ReplicaCount defines how many replicas deployment will scale into
	ReplicaCount int32 `json:"replicaCount,omitempty"`
}

// ScheduledScalerStatus defines the observed state of ScheduledScaler
type ScheduledScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// StoredReplicaCount store information original replicas
	StoredReplicaCount int32 `json:"storedReplicaCount,omitempty"`
	// Phase store information about phase of this resource
	Phase string `json:"phase,omitempty"`
	// LastVersion store information about last resource version before this reconcile
	LastVersion string `json:"lastVersion,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Deployment",type=string,JSONPath=`.spec.deploymentName`
// +kubebuilder:printcolumn:name="Orig Replicas",type=integer,JSONPath=`.status.storedReplicaCount`
// +kubebuilder:printcolumn:name="Target Replicas",type=integer,JSONPath=`.spec.replicaCount`

// ScheduledScaler is the Schema for the scheduledscalers API
type ScheduledScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledScalerSpec   `json:"spec,omitempty"`
	Status ScheduledScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScheduledScalerList contains a list of ScheduledScaler
type ScheduledScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledScaler{}, &ScheduledScalerList{})
}
