package v2

import (
	v1 "maistra.io/api/core/v1"
)

// StackdriverAddonConfig configuration specific to Stackdriver integration.
type StackdriverAddonConfig struct {
	// Configuration for Stackdriver tracer.  Applies when Addons.Tracer.Type=Stackdriver
	Tracer *StackdriverTracerConfig `json:"tracer,omitempty"`
	// Configuration for Stackdriver telemetry plugins.  Applies when telemetry
	// is enabled
	Telemetry *StackdriverTelemetryConfig `json:"telemetry,omitempty"`
}

// StackdriverTracerConfig configures the Stackdriver tracer
type StackdriverTracerConfig struct {
	// .Values.global.tracer.stackdriver.debug
	// +optional
	Debug *bool `json:"debug,omitempty"`
	// .Values.global.tracer.stackdriver.maxNumberOfAttributes
	// +optional
	MaxNumberOfAttributes *int64 `json:"maxNumberOfAttributes,omitempty"`
	// .Values.global.tracer.stackdriver.maxNumberOfAnnotations
	// +optional
	MaxNumberOfAnnotations *int64 `json:"maxNumberOfAnnotations,omitempty"`
	// .Values.global.tracer.stackdriver.maxNumberOfMessageEvents
	// +optional
	MaxNumberOfMessageEvents *int64 `json:"maxNumberOfMessageEvents,omitempty"`
}

// StackdriverTelemetryConfig adds telemetry filters for Stackdriver.
type StackdriverTelemetryConfig struct {
	// Enable installation of Stackdriver telemetry filters (mixer or v2/envoy).
	// These will only be installed if this is enabled an telemetry is enabled.
	Enablement `json:",inline"`
	// Auth configuration for stackdriver adapter (mixer/v1 telemetry only)
	// .Values.mixer.adapters.stackdriver.auth
	// +optional
	Auth *StackdriverAuthConfig `json:"auth,omitempty"`
	// EnableContextGraph for stackdriver adapter (edge reporting)
	// .Values.mixer.adapters.stackdriver.contextGraph.enabled, defaults to false
	// .Values.telemetry.v2.stackdriver.topology, defaults to false
	// +optional
	EnableContextGraph *bool `json:"enableContextGraph,omitempty"`
	// EnableLogging for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.logging.enabled, defaults to true
	// .Values.telemetry.v2.stackdriver.logging, defaults to false
	// +optional
	EnableLogging *bool `json:"enableLogging,omitempty"`
	// EnableMetrics for stackdriver adapter
	// .Values.mixer.adapters.stackdriver.metrics.enabled, defaults to true
	// .Values.telemetry.v2.stackdriver.monitoring??? defaults to false
	// +optional
	EnableMetrics *bool `json:"enableMetrics,omitempty"`
	// DisableOutbound disables intallation of sidecar outbound filter
	// .Values.telemetry.v2.stackdriver.disableOutbound, defaults to false
	// +optional
	//DisableOutbound bool `json:"disableOutbound,omitempty"`
	// AccessLogging configures access logging for stackdriver
	AccessLogging *StackdriverAccessLogTelemetryConfig `json:"accessLogging,omitempty"`
	//ConfigOverride apply custom configuration to Stackdriver filters (v2
	// telemetry only)
	// .Values.telemetry.v2.stackdriver.configOverride
	// +optional
	ConfigOverride *v1.HelmValues `json:"configOverride,omitempty"`
}

// StackdriverAuthConfig is the auth config for stackdriver.  Only one field may be set
type StackdriverAuthConfig struct {
	// AppCredentials if true, use default app credentials.
	// .Values.mixer.adapters.stackdriver.auth.appCredentials, defaults to false
	// +optional
	AppCredentials *bool `json:"appCredentials,omitempty"`
	// APIKey use the specified key.
	// .Values.mixer.adapters.stackdriver.auth.apiKey
	// +optional
	APIKey string `json:"apiKey,omitempty"`
	// ServiceAccountPath use the path to the service account.
	// .Values.mixer.adapters.stackdriver.auth.serviceAccountPath
	// +optional
	ServiceAccountPath string `json:"serviceAccountPath,omitempty"`
}

// StackdriverAccessLogTelemetryConfig for v2 telemetry.
type StackdriverAccessLogTelemetryConfig struct {
	// Enable installation of access log filter.
	// .Values.telemetry.v2.accessLogPolicy.enabled
	Enablement `json:",inline"`
	// LogWindowDuration configures the log window duration for access logs.
	// defaults to 43200s
	// To reduce the number of successful logs, default log window duration is
	// set to 12 hours.
	// .Values.telemetry.v2.accessLogPolicy.logWindowDuration
	// +optional
	LogWindowDuration string `json:"logWindowDuration,omitempty"`
}
