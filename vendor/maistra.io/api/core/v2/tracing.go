package v2

// TracingConfig configures tracing solutions for the mesh.
// .Values.global.enableTracing
type TracingConfig struct {
	// Type represents the type of tracer to be installed.
	Type TracerType `json:"type,omitempty"`
	// Sampling sets the mesh-wide trace sampling percentage. Should be between
	// 0.0 - 100.0. Precision to 0.01, scaled as 0 to 10000, e.g.: 100% = 10000,
	// 1% = 100
	// .Values.pilot.traceSampling
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	// +optional
	Sampling *int32 `json:"sampling,omitempty"`
}

// TracerType represents the tracer type to use
type TracerType string

const (
	// TracerTypeNone is used to represent no tracer
	TracerTypeNone TracerType = "None"
	// TracerTypeJaeger is used to represent Jaeger as the tracer
	TracerTypeJaeger TracerType = "Jaeger"
	// TracerTypeStackdriver is used to represent Stackdriver as the tracer
	TracerTypeStackdriver TracerType = "Stackdriver"
	// TracerTypeZipkin      TracerType = "Zipkin"
	// TracerTypeLightstep   TracerType = "Lightstep"
	// TracerTypeDatadog     TracerType = "Datadog"
)
