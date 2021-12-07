package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "maistra.io/api/core/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlane is the Schema for the controlplanes API
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=smcp,categories=maistra-io
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.annotations.readyComponentCount",description="How many of the total number of components are ready"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason",description="Whether or not the control plane installation is up to date and ready to handle requests."
// +kubebuilder:printcolumn:name="Profiles",type="string",JSONPath=".status.appliedSpec.profiles",description="The configuration profiles applied to the configuration."
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.chartVersion",description="The actual current version of the control plane installation."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the object"
// +kubebuilder:printcolumn:name="Image Registry",type="string",JSONPath=".status.appliedSpec.runtime.defaults.container.registry",description="The image registry used as the base for all component images.",priority=1
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
	// +optional
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlaneList contains a list of ServiceMeshControlPlane
type ServiceMeshControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshControlPlane `json:"items"`
}

// ControlPlaneStatus defines the observed state of ServiceMeshControlPlane
// ControlPlaneStatus represents the current state of a ServiceMeshControlPlane.
type ControlPlaneStatus struct {
	v1.StatusBase `json:",inline"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	v1.StatusType `json:",inline"`

	// The generation observed by the controller during the most recent
	// reconciliation. The information in the status pertains to this particular
	// generation of the object.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// The version of the operator that last processed this resource.
	OperatorVersion string `json:"operatorVersion,omitempty"`

	// The version of the charts that were last processed for this resource.
	ChartVersion string `json:"chartVersion,omitempty"`

	// The list of components comprising the control plane and their statuses.
	v1.ComponentStatusList `json:",inline"`

	// The readiness status of components & owned resources
	Readiness ReadinessStatus `json:"readiness"`

	// The resulting specification of the configuration options after all profiles
	// have been applied.
	// +optional
	AppliedSpec ControlPlaneSpec `json:"appliedSpec,omitempty"`

	// The resulting values.yaml used to generate the charts.
	// +optional
	AppliedValues v1.ControlPlaneSpec `json:"appliedValues,omitempty"`
}

// ReadinessStatus contains readiness information for each deployed component.
type ReadinessStatus struct {
	// The readiness status of components
	// +optional
	Components ReadinessMap `json:"components,omitempty"`
}

type ReadinessMap map[string][]string

// GetReconciledVersion returns the reconciled version, or a default for older resources
func (s *ControlPlaneStatus) GetReconciledVersion() string {
	if s == nil {
		return v1.ComposeReconciledVersion("0.0.0", 0)
	}
	return v1.ComposeReconciledVersion(s.OperatorVersion, s.ObservedGeneration)
}

// ControlPlaneSpec represents the configuration for installing a control plane
type ControlPlaneSpec struct {
	// XXX: the resource name is intended to be used as the revision name, which
	// is used by istio.io/rev labels/annotations to specify which control plane
	// workloads should be connecting with.

	// Profiles selects the profile to use for default values. Defaults to
	// "default" when not set.
	// +optional
	Profiles []string `json:"profiles,omitempty"`

	// Version specifies what Maistra version of the control plane to install.
	// When creating a new ServiceMeshControlPlane with an empty version, the
	// admission webhook sets the version to the current version.
	// Existing ServiceMeshControlPlanes with an empty version are treated as
	// having the version set to "v1.0"
	// +optional
	Version string `json:"version,omitempty"`
	// Cluster is the general configuration of the cluster (cluster name,
	// network name, multi-cluster, mesh expansion, etc.)
	// +optional
	Cluster *ControlPlaneClusterConfig `json:"cluster,omitempty"`
	// General represents general control plane configuration that does not
	// logically fit in another area.
	// +optional
	General *GeneralConfig `json:"general,omitempty"`
	// Policy configures policy checking for the control plane.
	// .Values.policy.enabled, true if not null
	// +optional
	Policy *PolicyConfig `json:"policy,omitempty"`
	// Proxy configures the default behavior for sidecars.  Many values were
	// previously exposed through .Values.global.proxy
	// +optional
	Proxy *ProxyConfig `json:"proxy,omitempty"`
	// Security configures aspects of security for the control plane.
	// +optional
	Security *SecurityConfig `json:"security,omitempty"`
	// Telemetry configures telemetry for the mesh.
	// .Values.mixer.telemetry.enabled, true if not null.  1.6, .Values.telemetry.enabled
	// +optional
	Telemetry *TelemetryConfig `json:"telemetry,omitempty"`
	// Tracing configures tracing for the mesh.
	// +optional
	Tracing *TracingConfig `json:"tracing,omitempty"`
	// Gateways configures gateways for the mesh
	// .Values.gateways.*
	// +optional
	Gateways *GatewaysConfig `json:"gateways,omitempty"`
	// Runtime configuration for pilot (and galley, etc., pre 2.0)
	// +optional
	Runtime *ControlPlaneRuntimeConfig `json:"runtime,omitempty"`
	// Addons is used to configure additional features beyond core control plane
	// components, e.g. visualization, metric storage, etc.
	// +optional
	Addons *AddonsConfig `json:"addons,omitempty"`
	// TechPreview contains switches for features that are not GA yet.
	// +optional
	TechPreview *v1.HelmValues `json:"techPreview,omitempty"`
}

// Enablement is a common definition for features that can be enabled
type Enablement struct {
	// Enabled specifies whether or not this feature is enabled
	Enabled *bool `json:"enabled,omitempty"`
}

func (s ControlPlaneSpec) IsKialiEnabled() bool {
	return s.Addons != nil &&
		s.Addons.Kiali != nil &&
		s.Addons.Kiali.Enabled != nil &&
		*s.Addons.Kiali.Enabled
}

func (s ControlPlaneSpec) IsPrometheusEnabled() bool {
	return s.Addons.Prometheus != nil &&
		s.Addons.Prometheus.Enabled != nil &&
		*s.Addons.Prometheus.Enabled
}
