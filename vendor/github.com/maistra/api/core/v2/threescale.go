package v2

// ThreeScaleAddonConfig represents configuration options for the installation of the
// 3scale adapter.  The options are structured similarly to what is defined by
// the 3scale ConfigMap.
type ThreeScaleAddonConfig struct {
	Enablement `json:",inline"`

	// ListenerAddr sets the listen address for the gRPC server.
	// PARAM_THREESCALE_LISTEN_ADDR
	// +optional
	ListenAddr *int32 `json:"listen_addr,omitempty"`
	// LogGRPC controls whether the log includes gRPC info
	// PARAM_THREESCALE_LOG_GRPC
	// +optional
	LogGRPC *bool `json:"log_grpc,omitempty"`
	// LogJSON controls whether the log is formatted as JSON
	// PARAM_THREESCALE_LOG_JSON
	// +optional
	LogJSON *bool `json:"log_json,omitempty"`
	// LogLevel sets the minimum log output level. Accepted values are one of:
	// debug, info, warn, error, none
	// PARAM_THREESCALE_LOG_LEVEL
	// +optional
	LogLevel string `json:"log_level,omitempty"`

	// Metrics configures metrics specific details
	// +optional
	Metrics *ThreeScaleMetricsConfig `json:"metrics,omitempty"`

	// System configures system specific details
	// +optional
	System *ThreeScaleSystemConfig `json:"system,omitempty"`

	// Client configures client specific details
	// +optional
	Client *ThreeScaleClientConfig `json:"client,omitempty"`

	// GRPC configures gRPC specific details
	// +optional
	GRPC *ThreeScaleGRPCConfig `json:"grpc,omitempty"`

	// Backend configures backend specific details
	// +optional
	Backend *ThreeScaleBackendConfig `json:"backend,omitempty"`
}

// ThreeScaleMetricsConfig represents 3scale adapter options for its 'metrics'
// section.
type ThreeScaleMetricsConfig struct {
	// Port sets the port which 3scale /metrics endpoint can be scrapped from
	// PARAM_THREESCALE_METRICS_PORT
	// +optional
	Port *int32 `json:"port,omitempty"`
	// Report controls whether 3scale system and backend metrics are collected
	// and reported to Prometheus
	// PARAM_THREESCALE_REPORT_METRICS
	// +optional
	Report *bool `json:"report,omitempty"`
}

// ThreeScaleSystemConfig represents 3scale adapter options for its 'system'
// section.
type ThreeScaleSystemConfig struct {
	// CacheMaxSize is the max number of items that can be stored in the cache
	// at any time. Set to 0 to disable caching
	// PARAM_THREESCALE_CACHE_ENTRIES_MAX
	// +optional
	CacheMaxSize *int64 `json:"cache_max_size,omitempty"`
	// CacheRefreshRetries sets the number of times unreachable hosts will be
	// retried during a cache update loop
	// PARAM_THREESCALE_CACHE_REFRESH_RETRIES
	// +optional
	CacheRefreshRetries *int32 `json:"cache_refresh_retries,omitempty"`
	// CacheRefreshInterval is the time period in seconds, before a background
	// process attempts to refresh cached entries
	// PARAM_THREESCALE_CACHE_REFRESH_SECONDS
	// +optional
	CacheRefreshInterval *int32 `json:"cache_refresh_interval,omitempty"`
	// CacheTTL is the time period, in seconds, to wait before purging expired
	// items from the cache
	// PARAM_THREESCALE_CACHE_TTL_SECONDS
	// +optional
	CacheTTL *int32 `json:"cache_ttl,omitempty"`
}

// ThreeScaleClientConfig represents 3scale adapter options for its 'client'
// section.
type ThreeScaleClientConfig struct {
	// AllowInsecureConnections skips certificate verification when calling
	// 3scale API's. Enabling is not recommended
	// PARAM_THREESCALE_ALLOW_INSECURE_CONN
	// +optional
	AllowInsecureConnections *bool `json:"allow_insecure_connections,omitempty"`
	// Timeout sets the number of seconds to wait before terminating requests
	// to 3scale System and Backend
	// PARAM_THREESCALE_CLIENT_TIMEOUT_SECONDS
	// +optional
	Timeout *int32 `json:"timeout,omitempty"`
}

// ThreeScaleGRPCConfig represents 3scale adapter options for its 'grpc'
// section.
type ThreeScaleGRPCConfig struct {
	// MaxConnTimeout sets the maximum amount of seconds (+/-10% jitter) a
	// connection may exist before it will be closed
	// PARAM_THREESCALE_GRPC_CONN_MAX_SECONDS
	// +optional
	MaxConnTimeout *int32 `json:"max_conn_timeout,omitempty"`
}

// ThreeScaleBackendConfig represents 3scale adapter options for its 'backend'
// section.
type ThreeScaleBackendConfig struct {
	// EnableCache if true, attempts to create an in-memory apisonator cache for
	// authorization requests
	// PARAM_THREESCALE_USE_CACHED_BACKEND
	// +optional
	EnableCache *bool `json:"enable_cache,omitempty"`
	// CacheFlushInterval sets the interval at which metrics get reported from
	// the cache to 3scale
	// PARAM_THREESCALE_BACKEND_CACHE_FLUSH_INTERVAL_SECONDS
	// +optional
	CacheFlushInterval *int32 `json:"cache_flush_interval,omitempty"`
	// PolicyFailClosed if true, request will fail if 3scale Apisonator is
	// unreachable
	// PARAM_THREESCALE_BACKEND_CACHE_POLICY_FAIL_CLOSED
	// +optional
	PolicyFailClosed *bool `json:"policy_fail_closed,omitempty"`
}
