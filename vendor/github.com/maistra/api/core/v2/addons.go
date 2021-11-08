package v2

// AddonsConfig configures additional features for use with the mesh
type AddonsConfig struct {
	// Prometheus configures Prometheus specific addon capabilities
	Prometheus *PrometheusAddonConfig `json:"prometheus,omitempty"`
	// Stackdriver configures Stackdriver specific addon capabilities
	Stackdriver *StackdriverAddonConfig `json:"stackdriver,omitempty"`
	// Jaeger configures Jaeger specific addon capabilities
	Jaeger *JaegerAddonConfig `json:"jaeger,omitempty"`
	// Grafana configures a grafana instance to use with the mesh
	// .Values.grafana.enabled, true if not null
	// +optional
	Grafana *GrafanaAddonConfig `json:"grafana,omitempty"`
	// Kiali configures a kiali instance to use with the mesh
	// .Values.kiali.enabled, true if not null
	// +optional
	Kiali *KialiAddonConfig `json:"kiali,omitempty"`
	// ThreeScale configures the 3scale adapter
	// +optional
	ThreeScale *ThreeScaleAddonConfig `json:"3scale,omitempty"`
}
