package v2

// PrometheusAddonConfig configures a prometheus instance to be used by the
// control plane.  Only one of Install or Address may be specified
type PrometheusAddonConfig struct {
	Enablement `json:",inline"`
	// MetricsExpiryDuration is the duration to hold metrics. (mixer/v1 only)
	// .Values.mixer.adapters.prometheus.metricsExpiryDuration, defaults to 10m
	// +optional
	MetricsExpiryDuration string `json:"metricsExpiryDuration,omitempty"`
	// Scrape metrics from the pod if true. (maistra-2.0+)
	// defaults to true
	// .Values.meshConfig.enablePrometheusMerge
	// +optional
	Scrape *bool `json:"scrape,omitempty"`
	// Install configuration if not using an existing prometheus installation.
	// .Values.prometheus.enabled, if not null
	// +optional
	Install *PrometheusInstallConfig `json:"install,omitempty"`
	// Address of existing prometheus installation
	// implies .Values.kiali.prometheusAddr
	// XXX: do we need to do anything to configure credentials for accessing
	// the prometheus server?
	// +optional
	Address *string `json:"address,omitempty"`
}

// PrometheusInstallConfig represents the configuration to be applied when
// installing a new instance of prometheus for use with the mesh.
type PrometheusInstallConfig struct {
	// SelfManaged specifies whether or not the entire install should be managed
	// by Maistra (true) or the Prometheus operator (false, not supported).
	// Governs use of either prometheus charts or prometheusOperator charts.
	// +optional
	SelfManaged bool `json:"selfManaged,omitempty"`
	// Retention specifies how long metrics should be retained by prometheus.
	// .Values.prometheus.retention, defaults to 6h
	// +optional
	Retention string `json:"retention,omitempty"`
	// ScrapeInterval specifies how frequently prometheus should scrape pods for
	// metrics.
	// .Values.prometheus.scrapeInterval, defaults to 15s
	// +optional
	ScrapeInterval string `json:"scrapeInterval,omitempty"`
	// Service allows for customization of the k8s Service associated with the
	// prometheus installation.
	// +optional
	Service *ComponentServiceConfig `json:"service,omitempty"`
	// UseTLS for the prometheus server
	// .Values.prometheus.provisionPrometheusCert
	// 1.6+
	//ProvisionCert bool
	// this seems to overlap with provision cert, as this manifests something similar to the above
	// .Values.prometheus.security.enabled, version < 1.6
	//EnableSecurity bool
	// +optional
	UseTLS *bool `json:"useTLS,omitempty"`
}
