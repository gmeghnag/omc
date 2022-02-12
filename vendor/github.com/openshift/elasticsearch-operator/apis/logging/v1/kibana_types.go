package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Specification of the desired behavior of the Kibana
//
// +k8s:openapi-gen=true
type KibanaSpec struct {
	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator
	//
	ManagementState ManagementState `json:"managementState"`

	// The resource requirements for the Kibana nodes
	//
	// +nullable
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Resource Requirements"
	Resources *corev1.ResourceRequirements `json:"resources"`

	// The node selector to use for the Kibana Visualization component
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:nodeSelector"
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// The desired number of Kibana Pods for the Visualization component
	//
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kibana Size",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	Replicas int32 `json:"replicas"`

	// Specification of the Kibana Proxy component
	//
	// +optional
	ProxySpec `json:"proxy,omitempty"`
}

type ProxySpec struct {
	// The resource requirements for Kibana proxy
	//
	// +nullable
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources"`
}

// KibanaStatus defines the observed state of Kibana
// +k8s:openapi-gen=true
type KibanaStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Deployment string `json:"deployment"`
	// +optional
	ReplicaSets []string `json:"replicaSets,omitempty"`
	// The status for each of the Kibana pods for the Visualization component
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Kibana Status",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Pods PodStateMap `json:"pods,omitempty"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kibanas,categories=logging,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +kubebuilder:printcolumn:name="Replicas",JSONPath=".spec.replicas",type=integer
// Kibana instance
// +operator-sdk:csv:customresourcedefinitions:displayName="Kibana",resources={{Deployment,v1},{ConsoleExternalLogLink,v1},{ConsoleLink,v1},{ConfigMap,v1},{Role,v1},{RoleBinding,v1},{Route,v1},{Service,v1},{ServiceAccount,v1}}
type Kibana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KibanaSpec `json:"spec,omitempty"`

	Status []KibanaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KibanaList contains a list of Kibana
type KibanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kibana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kibana{}, &KibanaList{})
}
