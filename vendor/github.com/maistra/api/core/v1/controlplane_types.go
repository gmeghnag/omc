package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultTemplate = "default"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlane represents a deployment of the service mesh control
// plane. The control plane components are deployed in the namespace in which
// the ServiceMeshControlPlane resides. The configuration options for the
// components that comprise the control plane are specified in this object.
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=smcp,categories=maistra-io
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.annotations.readyComponentCount",description="How many of the total number of components are ready"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Reconciled\")].reason",description="Whether or not the control plane installation is up to date."
// +kubebuilder:printcolumn:name="Template",type="string",JSONPath=".status.lastAppliedConfiguration.template",description="The configuration template to use as the base."
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.lastAppliedConfiguration.version",description="The actual current version of the control plane installation."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the object"
// +kubebuilder:printcolumn:name="Image HUB",type="string",JSONPath=".status.lastAppliedConfiguration.istio.global.hub",description="The image hub used as the base for all component images.",priority=1
type ServiceMeshControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specification of the desired state of this ServiceMeshControlPlane.
	// This includes the configuration options for all components that comprise
	// the control plane.
	// +kubebuilder:validation:Required
	Spec ControlPlaneSpec `json:"spec"`

	// The current status of this ServiceMeshControlPlane and the components
	// that comprise the control plane. This data may be out of date by some
	// window of time.
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlaneList contains a list of ServiceMeshControlPlane
type ServiceMeshControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshControlPlane `json:"items"`
}

// ControlPlaneStatus represents the current state of a ServiceMeshControlPlane.
type ControlPlaneStatus struct {
	StatusBase `json:",inline"`

	StatusType `json:",inline"`

	// The generation observed by the controller during the most recent
	// reconciliation. The information in the status pertains to this particular
	// generation of the object.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// The last version that was reconciled.
	ReconciledVersion string `json:"reconciledVersion,omitempty"`

	// The list of components comprising the control plane and their statuses.
	// +nullable
	ComponentStatusList `json:",inline"`

	// The full specification of the configuration options that were applied
	// to the components of the control plane during the most recent reconciliation.
	// +optional
	LastAppliedConfiguration ControlPlaneSpec `json:"lastAppliedConfiguration"`
}

// GetReconciledVersion returns the reconciled version, or a default for older resources
func (s *ControlPlaneStatus) GetReconciledVersion() string {
	if s == nil {
		return ComposeReconciledVersion("0.0.0", 0)
	}
	if s.ReconciledVersion == "" {
		return ComposeReconciledVersion("1.0.0", s.ObservedGeneration)
	}
	return s.ReconciledVersion
}

// ControlPlaneSpec represents the configuration for installing a control plane.
type ControlPlaneSpec struct {
	// Template selects the template to use for default values. Defaults to
	// "default" when not set.
	// DEPRECATED - use Profiles instead
	// +optional
	Template string `json:"template,omitempty"`

	// Profiles selects the profile to use for default values. Defaults to
	// "default" when not set.  Takes precedence over Template.
	// +optional
	Profiles []string `json:"profiles,omitempty"`

	// Version specifies what Maistra version of the control plane to install.
	// When creating a new ServiceMeshControlPlane with an empty version, the
	// admission webhook sets the version to the latest version supported by
	// the operator.
	// Existing ServiceMeshControlPlanes with an empty version are treated as
	// having the version set to "v1.0"
	// +optional
	Version string `json:"version,omitempty"`

	// DEPRECATED: No longer used anywhere.
	// Previously used to specify the NetworkType of the cluster. Defaults to "subnet".
	// +optional
	NetworkType NetworkType `json:"networkType,omitempty"`

	// Specifies the Istio configuration options that are passed to Helm when the
	// Istio charts are rendered. These options are usually populated from the
	// template specified in the spec.template field, but individual values can
	// be overridden here.
	// More info: https://maistra.io/docs/installation/installation-options/
	// +optional
	// +kubebuilder:validation:Optional
	Istio *HelmValues `json:"istio,omitempty"`

	// Specifies the 3Scale configuration options that are passed to Helm when the
	// 3Scale charts are rendered. These values are usually populated from the
	// template specified in the spec.template field, but individual values can
	// be overridden here.
	// More info: https://maistra.io/docs/installation/installation-options/#_3scale
	// +optional
	// +kubebuilder:validation:Optional
	ThreeScale *HelmValues `json:"threeScale,omitempty"`
}

// NetworkType is type definition representing the network type of the cluster
type NetworkType string

const (
	// NetworkTypeSubnet when using ovs-subnet
	NetworkTypeSubnet NetworkType = "subnet"
	// NetworkTypeMultitenant when using ovs-multitenant
	NetworkTypeMultitenant NetworkType = "multitenant"
	// NetworkTypeNetworkPolicy when using ovs-networkpolicy
	NetworkTypeNetworkPolicy NetworkType = "networkpolicy"
)
