package v2

// PolicyConfig configures policy aspects of the mesh.
type PolicyConfig struct {
	// Required, the policy implementation
	// defaults to Istiod 1.6+, Mixer pre-1.6
	Type PolicyType `json:"type,omitempty"`
	// Mixer configuration (legacy, v1)
	// .Values.mixer.policy.enabled
	// +optional
	Mixer *MixerPolicyConfig `json:"mixer,omitempty"`
	// Remote mixer configuration (legacy, v1)
	// .Values.global.remotePolicyAddress
	// +optional
	Remote *RemotePolicyConfig `json:"remote,omitempty"`
}

// PolicyType represents the type of policy implementation used by the mesh.
type PolicyType string

const (
	// PolicyTypeNone represents disabling of policy
	// XXX: note, this doesn't appear to affect Istio 1.6, i.e. no different than Istiod setting
	PolicyTypeNone PolicyType = "None"
	// PolicyTypeMixer represents mixer, v1 implementation
	PolicyTypeMixer PolicyType = "Mixer"
	// PolicyTypeRemote represents remote mixer, v1 implementation
	PolicyTypeRemote PolicyType = "Remote"
	// PolicyTypeIstiod represents istio, v2 implementation
	PolicyTypeIstiod PolicyType = "Istiod"
)

// MixerPolicyConfig configures a mixer implementation for policy
// .Values.mixer.policy.enabled
type MixerPolicyConfig struct {
	// EnableChecks configures whether or not policy checks should be enabled.
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	// +optional
	EnableChecks *bool `json:"enableChecks,omitempty"`
	// FailOpen configures policy checks to fail if mixer cannot be reached.
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	// +optional
	FailOpen *bool `json:"failOpen,omitempty"`
	// SessionAffinity configures session affinity for sidecar policy connections.
	// .Values.mixer.policy.sessionAffinityEnabled
	// +optional
	SessionAffinity *bool `json:"sessionAffinity,omitempty"`
	// Adapters configures available adapters.
	// +optional
	Adapters *MixerPolicyAdaptersConfig `json:"adapters,omitempty"`
}

// MixerPolicyAdaptersConfig configures policy adapters for mixer.
type MixerPolicyAdaptersConfig struct {
	// UseAdapterCRDs configures mixer to support deprecated mixer CRDs.
	// .Values.mixer.policy.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	// Only supported in v1.0, where it defaulted to true
	// +optional
	UseAdapterCRDs *bool `json:"useAdapterCRDs,omitempty"`
	// Kubernetesenv configures the use of the kubernetesenv adapter.
	// .Values.mixer.policy.adapters.kubernetesenv.enabled, defaults to true
	// +optional
	KubernetesEnv *bool `json:"kubernetesenv,omitempty"`
}

// RemotePolicyConfig configures a remote mixer instance for policy
type RemotePolicyConfig struct {
	// Address represents the address of the mixer server.
	// .Values.global.remotePolicyAddress, maps to MeshConfig.mixerCheckServer
	Address string `json:"address,omitempty"`
	// CreateServices specifies whether or not a k8s Service should be created for the remote policy server.
	// .Values.global.createRemoteSvcEndpoints
	// +optional
	CreateService *bool `json:"createService,omitempty"`
	// EnableChecks configures whether or not policy checks should be enabled.
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	// +optional
	EnableChecks *bool `json:"enableChecks,omitempty"`
	// FailOpen configures policy checks to fail if mixer cannot be reached.
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	// +optional
	FailOpen *bool `json:"failOpen,omitempty"`
}
