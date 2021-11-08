package v2

// TelemetryConfig for the mesh
type TelemetryConfig struct {
	// Type of telemetry implementation to use.
	Type TelemetryType `json:"type,omitempty"`
	// Mixer represents legacy, v1 telemetry.
	// implies .Values.telemetry.v1.enabled, if not null
	// +optional
	Mixer *MixerTelemetryConfig `json:"mixer,omitempty"`
	// Remote represents a remote, legacy, v1 telemetry.
	// +optional
	Remote *RemoteTelemetryConfig `json:"remote,omitempty"`
}

// TelemetryType represents the telemetry implementation used.
type TelemetryType string

const (
	// TelemetryTypeNone disables telemetry
	TelemetryTypeNone TelemetryType = "None"
	// TelemetryTypeMixer represents mixer telemetry, v1
	TelemetryTypeMixer TelemetryType = "Mixer"
	// TelemetryTypeRemote represents remote mixer telemetry server, v1
	TelemetryTypeRemote TelemetryType = "Remote"
	// TelemetryTypeIstiod represents istio, v2
	TelemetryTypeIstiod TelemetryType = "Istiod"
)

// MixerTelemetryConfig is the configuration for legacy, v1 mixer telemetry.
// .Values.telemetry.v1.enabled
type MixerTelemetryConfig struct {
	// SessionAffinity configures session affinity for sidecar telemetry connections.
	// .Values.mixer.telemetry.sessionAffinityEnabled, maps to MeshConfig.sidecarToTelemetrySessionAffinity
	// +optional
	SessionAffinity *bool `json:"sessionAffinity,omitempty"`
	// Loadshedding configuration for telemetry
	// .Values.mixer.telemetry.loadshedding
	// +optional
	Loadshedding *TelemetryLoadSheddingConfig `json:"loadshedding,omitempty"`
	// Batching settings used when sending telemetry.
	// +optional
	Batching *TelemetryBatchingConfig `json:"batching,omitempty"`
	// Adapters configures the adapters used by mixer telemetry.
	// +optional
	Adapters *MixerTelemetryAdaptersConfig `json:"adapters,omitempty"`
}

// TelemetryLoadSheddingConfig configures how mixer telemetry loadshedding behaves
type TelemetryLoadSheddingConfig struct {
	// Mode represents the loadshedding mode applied to mixer when it becomes
	// overloaded.  Valid values: disabled, logonly or enforce
	// .Values.mixer.telemetry.loadshedding.mode
	// +optional
	Mode string `json:"mode,omitempty"`
	// LatencyThreshold --
	// .Values.mixer.telemetry.loadshedding.latencyThreshold
	// +optional
	LatencyThreshold string `json:"latencyThreshold,omitempty"`
}

// TelemetryBatchingConfig configures how telemetry data is batched.
type TelemetryBatchingConfig struct {
	// MaxEntries represents the maximum number of entries to collect before sending them to mixer.
	// .Values.mixer.telemetry.reportBatchMaxEntries, maps to MeshConfig.reportBatchMaxEntries
	// Set reportBatchMaxEntries to 0 to use the default batching behavior (i.e., every 100 requests).
	// A positive value indicates the number of requests that are batched before telemetry data
	// is sent to the mixer server
	// +optional
	MaxEntries *int32 `json:"maxEntries,omitempty"`
	// MaxTime represents the maximum amount of time to hold entries before sending them to mixer.
	// .Values.mixer.telemetry.reportBatchMaxTime, maps to MeshConfig.reportBatchMaxTime
	// Set reportBatchMaxTime to 0 to use the default batching behavior (i.e., every 1 second).
	// A positive time value indicates the maximum wait time since the last request will telemetry data
	// be batched before being sent to the mixer server
	// +optional
	MaxTime string `json:"maxTime,omitempty"`
}

// MixerTelemetryAdaptersConfig is the configuration for mixer telemetry adapters.
type MixerTelemetryAdaptersConfig struct {
	// UseAdapterCRDs specifies whether or not mixer should support deprecated CRDs.
	// .Values.mixer.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	// XXX: i think this can be removed completely
	// +optional
	UseAdapterCRDs *bool `json:"useAdapterCRDs,omitempty"`
	// KubernetesEnv enables support for the kubernetesenv adapter.
	// .Values.mixer.adapters.kubernetesenv.enabled, defaults to true
	// +optional
	KubernetesEnv *bool `json:"kubernetesenv,omitempty"`
	// Stdio enables and configures the stdio adapter.
	// +optional
	Stdio *MixerTelemetryStdioConfig `json:"stdio,omitempty"`
}

// MixerTelemetryStdioConfig configures the stdio adapter for mixer telemetry.
type MixerTelemetryStdioConfig struct {
	// .Values.mixer.adapters.stdio.enabled
	Enablement `json:",inline"`
	// OutputAsJSON if true.
	// .Values.mixer.adapters.stdio.outputAsJson, defaults to false
	// +optional
	OutputAsJSON *bool `json:"outputAsJSON,omitempty"`
}

// RemoteTelemetryConfig configures a remote, legacy, v1 mixer telemetry.
// .Values.telemetry.v1.enabled true
type RemoteTelemetryConfig struct {
	// Address is the address of the remote telemetry server
	// .Values.global.remoteTelemetryAddress, maps to MeshConfig.mixerReportServer
	Address string `json:"address,omitempty"`
	// CreateService for the remote server.
	// .Values.global.createRemoteSvcEndpoints
	// +optional
	CreateService *bool `json:"createService,omitempty"`
	// Batching settings used when sending telemetry.
	// +optional
	Batching *TelemetryBatchingConfig `json:"batching,omitempty"`
}
