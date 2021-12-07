package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxyConfig configures the default sidecar behavior for workloads.
type ProxyConfig struct {
	// Logging configures logging for the sidecar.
	// e.g. .Values.global.proxy.logLevel
	// +optional
	Logging *ProxyLoggingConfig `json:"logging,omitempty"`
	// Networking represents network settings to be configured for the sidecars.
	// +optional
	Networking *ProxyNetworkingConfig `json:"networking,omitempty"`
	// Runtime is used to customize runtime configuration for the sidecar container.
	// +optional
	Runtime *ProxyRuntimeConfig `json:"runtime,omitempty"`
	// Injection is used to customize sidecar injection for the mesh.
	// +optional
	Injection *ProxyInjectionConfig `json:"injection,omitempty"`
	// AdminPort configures the admin port exposed by the sidecar.
	// maps to defaultConfig.proxyAdminPort, defaults to 15000
	// XXX: currently not configurable in charts
	// +optional
	AdminPort int32 `json:"adminPort,omitempty"`
	// Concurrency configures the number of threads that should be run by the sidecar.
	// .Values.global.proxy.concurrency, maps to defaultConfig.concurrency
	// XXX: removed in 1.7
	// XXX: this is defaulted to 2 in our values.yaml, but should probably be 0
	// +optional
	Concurrency *int32 `json:"concurrency,omitempty"`
	// AccessLogging configures access logging for proxies.
	// +optional
	AccessLogging *ProxyAccessLoggingConfig `json:"accessLogging,omitempty"`
	// EnvoyMetricsService configures reporting of Envoy metrics to an external
	// service.
	// .Values.global.proxy.envoyMetricsService
	// +optional
	EnvoyMetricsService *ProxyEnvoyServiceConfig `json:"envoyMetricsService,omitempty"`
}

// ProxyNetworkingConfig is used to configure networking aspects of the sidecar.
type ProxyNetworkingConfig struct {
	// ClusterDomain represents the domain for the cluster, defaults to cluster.local
	// .Values.global.proxy.clusterDomain
	// +optional
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// maps to meshConfig.defaultConfig.connectionTimeout, defaults to 10s
	// XXX: currently not exposed through values.yaml
	// +optional
	ConnectionTimeout string `json:"connectionTimeout,omitempty"`
	// MaxConnectionAge limits how long a sidecar can be connected to pilot.
	// This may be used to balance load across pilot instances, at the cost of
	// system churn.
	// .Values.pilot.keepaliveMaxServerConnectionAge
	// +optional
	MaxConnectionAge string `json:"maxConnectionAge,omitempty"`
	// Initialization is used to specify how the pod's networking through the
	// proxy is initialized.  This configures the use of CNI or an init container.
	// +optional
	Initialization *ProxyNetworkInitConfig `json:"initialization,omitempty"`
	// TrafficControl configures what network traffic is routed through the proxy.
	// +optional
	TrafficControl *ProxyTrafficControlConfig `json:"trafficControl,omitempty"`
	// Protocol configures how the sidecar works with applicaiton protocols.
	// +optional
	Protocol *ProxyNetworkProtocolConfig `json:"protocol,omitempty"`
	// DNS configures aspects of the sidecar's usage of DNS
	// +optional
	DNS *ProxyDNSConfig `json:"dns,omitempty"`
}

// ProxyNetworkInitConfig is used to configure how the pod's networking through
// the proxy is initialized.
type ProxyNetworkInitConfig struct {
	// Type of the network initialization implementation.
	Type ProxyNetworkInitType `json:"type,omitempty"`
	// InitContainer configures the use of a pod init container for initializing
	// the pod's networking.
	// istio_cni.enabled = false, if InitContainer is used
	// +optional
	InitContainer *ProxyInitContainerConfig `json:"initContainer,omitempty"`
}

// ProxyNetworkInitType represents the type of initializer to use for network initialization
type ProxyNetworkInitType string

const (
	// ProxyNetworkInitTypeCNI to use CNI for network initialization
	ProxyNetworkInitTypeCNI ProxyNetworkInitType = "CNI"
	// ProxyNetworkInitTypeInitContainer to use an init container for network initialization
	ProxyNetworkInitTypeInitContainer ProxyNetworkInitType = "InitContainer"
)

// ProxyInitContainerConfig configures execution aspects for the init container
type ProxyInitContainerConfig struct {
	// Runtime configures customization of the init container (e.g. resources)
	// +optional
	Runtime *ContainerConfig `json:"runtime,omitempty"`
}

// ProxyTrafficControlConfig configures what and how traffic is routed through
// the sidecar.
type ProxyTrafficControlConfig struct {
	// Inbound configures what inbound traffic is routed through the sidecar
	// traffic.sidecar.istio.io/includeInboundPorts defaults to * (all ports)
	// +optional
	Inbound ProxyInboundTrafficControlConfig `json:"inbound,omitempty"`
	// Outbound configures what outbound traffic is routed through the sidecar.
	// +optional
	Outbound ProxyOutboundTrafficControlConfig `json:"outbound,omitempty"`
}

// ProxyNetworkInterceptionMode represents the InterceptMode types.
type ProxyNetworkInterceptionMode string

const (
	// ProxyNetworkInterceptionModeRedirect requests iptables use REDIRECT to route inbound traffic through the sidecar.
	ProxyNetworkInterceptionModeRedirect ProxyNetworkInterceptionMode = "REDIRECT"
	// ProxyNetworkInterceptionModeTProxy requests iptables use TPROXY to route inbound traffic through the sidecar.
	ProxyNetworkInterceptionModeTProxy ProxyNetworkInterceptionMode = "TPROXY"
)

// ProxyInboundTrafficControlConfig configures what inbound traffic is
// routed through the sidecar.
type ProxyInboundTrafficControlConfig struct {
	// InterceptionMode specifies how traffic is directed through the sidecar.
	// maps to meshConfig.defaultConfig.interceptionMode, overridden by sidecar.istio.io/interceptionMode
	// XXX: currently not configurable through values.yaml
	// +optional
	InterceptionMode ProxyNetworkInterceptionMode `json:"interceptionMode,omitempty"`
	// IncludedPorts to be routed through the sidecar. * or comma separated list of integers
	// .Values.global.proxy.includeInboundPorts, defaults to * (all ports), overridden by traffic.sidecar.istio.io/includeInboundPorts
	// +optional
	IncludedPorts []string `json:"includedPorts,omitempty"`
	// ExcludedPorts to be routed around the sidecar.
	// .Values.global.proxy.excludeInboundPorts, defaults to empty list, overridden by traffic.sidecar.istio.io/excludeInboundPorts
	// +optional
	ExcludedPorts []int32 `json:"excludedPorts,omitempty"`
}

// ProxyOutboundTrafficControlConfig configure what outbound traffic is routed
// through the sidecar
type ProxyOutboundTrafficControlConfig struct {
	// IncludedIPRanges specifies which outbound IP ranges should be routed through the sidecar.
	// .Values.global.proxy.includeIPRanges, overridden by traffic.sidecar.istio.io/includeOutboundIPRanges
	// * or comma separated list of CIDR
	// +optional
	IncludedIPRanges []string `json:"includedIPRanges,omitempty"`
	// ExcludedIPRanges specifies which outbound IP ranges should _not_ be routed through the sidecar.
	// .Values.global.proxy.excludeIPRanges, overridden by traffic.sidecar.istio.io/excludeOutboundIPRanges
	// * or comma separated list of CIDR
	// +optional
	ExcludedIPRanges []string `json:"excludedIPRanges,omitempty"`
	// ExcludedPorts specifies which outbound ports should _not_ be routed through the sidecar.
	// .Values.global.proxy.excludeOutboundPorts, overridden by traffic.sidecar.istio.io/excludeOutboundPorts
	// comma separated list of integers
	// +optional
	ExcludedPorts []int32 `json:"excludedPorts,omitempty"`
	// Policy specifies what outbound traffic is allowed through the sidecar.
	// .Values.global.outboundTrafficPolicy.mode
	// +optional
	Policy ProxyOutboundTrafficPolicy `json:"policy,omitempty"`
}

// ProxyOutboundTrafficPolicy represents the outbound traffic policy type.
type ProxyOutboundTrafficPolicy string

const (
	// ProxyOutboundTrafficPolicyAllowAny allows all traffic through the sidecar.
	ProxyOutboundTrafficPolicyAllowAny ProxyOutboundTrafficPolicy = "ALLOW_ANY"
	// ProxyOutboundTrafficPolicyRegistryOnly only allows traffic destined for a
	// service in the service registry through the sidecar.  This limits outbound
	// traffic to only other services in the mesh.
	ProxyOutboundTrafficPolicyRegistryOnly ProxyOutboundTrafficPolicy = "REGISTRY_ONLY"
)

// ProxyNetworkProtocolConfig configures the sidecar's protocol handling.
type ProxyNetworkProtocolConfig struct {
	// AutoDetect configures automatic detection of connection protocols.
	// +optional
	AutoDetect *ProxyNetworkAutoProtocolDetectionConfig `json:"autoDetect,omitempty"`
}

// ProxyNetworkAutoProtocolDetectionConfig configures automatic protocol detection for the proxies.
type ProxyNetworkAutoProtocolDetectionConfig struct {
	// DetectionTimeout specifies how much time the sidecar will spend determining
	// the protocol being used for the connection before reverting to raw TCP.
	// .Values.global.proxy.protocolDetectionTimeout, maps to protocolDetectionTimeout
	// +optional
	Timeout string `json:"timeout,omitempty"`
	// EnableInboundSniffing enables protocol sniffing on inbound traffic.
	// .Values.pilot.enableProtocolSniffingForInbound
	// only supported for v1.1
	// +optional
	Inbound *bool `json:"inbound,omitempty"`
	// EnableOutboundSniffing enables protocol sniffing on outbound traffic.
	// .Values.pilot.enableProtocolSniffingForOutbound
	// only supported for v1.1
	// +optional
	Outbound *bool `json:"outbound,omitempty"`
}

// ProxyDNSConfig is used to configure aspects of the sidecar's DNS usage.
type ProxyDNSConfig struct {
	// SearchSuffixes are additional search suffixes to be used when resolving
	// names.
	// .Values.global.podDNSSearchNamespaces
	// Custom DNS config for the pod to resolve names of services in other
	// clusters. Use this to add additional search domains, and other settings.
	// see
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#dns-config
	// This does not apply to gateway pods as they typically need a different
	// set of DNS settings than the normal application pods (e.g., in
	// multicluster scenarios).
	// NOTE: If using templates, follow the pattern in the commented example below.
	//    podDNSSearchNamespaces:
	//    - global
	//    - "{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global"
	// +optional
	SearchSuffixes []string `json:"searchSuffixes,omitempty"`
	// RefreshRate configures the DNS refresh rate for Envoy cluster of type STRICT_DNS
	// This must be given it terms of seconds. For example, 300s is valid but 5m is invalid.
	// .Values.global.proxy.dnsRefreshRate, default 300s
	// +optional
	RefreshRate string `json:"refreshRate,omitempty"`
}

// ProxyRuntimeConfig customizes the runtime parameters of the sidecar container.
type ProxyRuntimeConfig struct {
	// Readiness configures the readiness probe behavior for the injected pod.
	// +optional
	Readiness *ProxyReadinessConfig `json:"readiness,omitempty"`
	// Container configures the sidecar container.
	// +optional
	Container *ContainerConfig `json:"container,omitempty"`
}

// ProxyReadinessConfig configures the readiness probe for the sidecar.
type ProxyReadinessConfig struct {
	// RewriteApplicationProbes specifies whether or not the injector should
	// rewrite application container probes to be routed through the sidecar.
	// .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe, defaults to false
	// rewrite probes for application pods to route through sidecar
	// +optional
	RewriteApplicationProbes bool `json:"rewriteApplicationProbes,omitempty"`
	// StatusPort specifies the port number to be used for status.
	// .Values.global.proxy.statusPort, overridden by status.sidecar.istio.io/port, defaults to 15020
	// Default port for Pilot agent health checks. A value of 0 will disable health checking.
	// XXX: this has no affect on which port is actually used for status.
	// +optional
	StatusPort int32 `json:"statusPort,omitempty"`
	// InitialDelaySeconds specifies the initial delay for the readiness probe
	// .Values.global.proxy.readinessInitialDelaySeconds, overridden by readiness.status.sidecar.istio.io/initialDelaySeconds, defaults to 1
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// PeriodSeconds specifies the period over which the probe is checked.
	// .Values.global.proxy.readinessPeriodSeconds, overridden by readiness.status.sidecar.istio.io/periodSeconds, defaults to 2
	// +optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// FailureThreshold represents the number of consecutive failures before the container is marked as not ready.
	// .Values.global.proxy.readinessFailureThreshold, overridden by readiness.status.sidecar.istio.io/failureThreshold, defaults to 30
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

// ProxyInjectionConfig configures sidecar injection for the mesh.
type ProxyInjectionConfig struct {
	// AutoInject configures automatic injection of sidecar proxies
	// .Values.global.proxy.autoInject
	// .Values.sidecarInjectorWebhook.enableNamespacesByDefault
	// +optional
	AutoInject *bool `json:"autoInject,omitempty"`
	// AlwaysInjectSelector allows specification of a label selector that when
	// matched will always inject a sidecar into the pod.
	// .Values.sidecarInjectorWebhook.alwaysInjectSelector
	// +optional
	AlwaysInjectSelector []metav1.LabelSelector `json:"alwaysInjectSelector,omitempty"`
	// NeverInjectSelector allows specification of a label selector that when
	// matched will never inject a sidecar into the pod.  This takes precendence
	// over AlwaysInjectSelector.
	// .Values.sidecarInjectorWebhook.neverInjectSelector
	// +optional
	NeverInjectSelector []metav1.LabelSelector `json:"neverInjectSelector,omitempty"`
	// InjectedAnnotations allows specification of additional annotations to be
	// added to pods that have sidecars injected in them.
	// .Values.sidecarInjectorWebhook.injectedAnnotations
	// +optional
	InjectedAnnotations map[string]string `json:"injectedAnnotations,omitempty"`
}

// ProxyAccessLoggingConfig configures access logging for proxies.  Multiple
// access logs can be configured.
type ProxyAccessLoggingConfig struct {
	// File configures access logging to the file system
	// +optional
	File *ProxyFileAccessLogConfig `json:"file,omitempty"`
	// File configures access logging to an envoy service
	// .Values.global.proxy.envoyAccessLogService
	// +optional
	EnvoyService *ProxyEnvoyServiceConfig `json:"envoyService,omitempty"`
}

// ProxyFileAccessLogConfig configures details related to file access log
type ProxyFileAccessLogConfig struct {
	// Name is the name of the file to which access log entries will be written.
	// If Name is not specified, no log entries will be written to a file.
	// .Values.global.proxy.accessLogFile
	// +optional
	Name string `json:"name,omitempty"`
	// Encoding to use when writing access log entries.  Currently, JSON or TEXT
	// may be specified.
	// .Values.global.proxy.accessLogEncoding
	// +optional
	Encoding string `json:"encoding,omitempty"`
	// Format to use when writing access log entries.
	// .Values.global.proxy.accessLogFormat
	// +optional
	Format string `json:"format,omitempty"`
}

// ProxyEnvoyServiceConfig configures reporting to an external Envoy service.
type ProxyEnvoyServiceConfig struct {
	// Enable sending Envoy metrics to the service.
	// .Values.global.proxy.(envoyAccessLogService|envoyMetricsService).enabled
	Enablement `json:",inline"`
	// Address of the service specified as host:port.
	// .Values.global.proxy.(envoyAccessLogService|envoyMetricsService).host
	// .Values.global.proxy.(envoyAccessLogService|envoyMetricsService).port
	// +optional
	Address string `json:"address,omitempty"`
	// TCPKeepalive configures keepalive settings to use when connecting to the
	// service.
	// .Values.global.proxy.(envoyAccessLogService|envoyMetricsService).tcpKeepalive
	// +optional
	TCPKeepalive *EnvoyServiceTCPKeepalive `json:"tcpKeepalive,omitempty"`
	// TLSSettings configures TLS settings to use when connecting to the service.
	// .Values.global.proxy.(envoyAccessLogService|envoyMetricsService).tlsSettings
	// +optional
	TLSSettings *EnvoyServiceClientTLSSettings `json:"tlsSettings,omitempty"`
}

// EnvoyServiceTCPKeepalive configures keepalive settings for the Envoy service.
// Provides the same interface as networking.v1alpha3.istio.io, ConnectionPoolSettings_TCPSettings_TcpKeepalive
type EnvoyServiceTCPKeepalive struct {
	// Probes represents the number of successive probe failures after which the
	// connection should be considered "dead."
	// +optional
	Probes uint32 `json:"probes,omitempty"`
	// Time represents the length of idle time that must elapse before a probe
	// is sent.
	// +optional
	Time string `json:"time,omitempty"`
	// Interval represents the interval between probes.
	// +optional
	Interval string `json:"interval,omitempty"`
}

// EnvoyServiceClientTLSSettings configures TLS settings for the Envoy service.
// Provides the same interface as networking.v1alpha3.istio.io, ClientTLSSettings
type EnvoyServiceClientTLSSettings struct {
	// Mode represents the TLS mode to apply to the connection.  The following
	// values are supported: DISABLE, SIMPLE, MUTUAL, ISTIO_MUTUAL
	// +optional
	Mode string `json:"mode,omitempty"`
	// ClientCertificate represents the file name containing the client certificate
	// to show to the Envoy service, e.g. /etc/istio/als/cert-chain.pem
	// +optional
	ClientCertificate string `json:"clientCertificate,omitempty"`
	// PrivateKey represents the file name containing the private key used by
	// the client, e.g. /etc/istio/als/key.pem
	// +optional
	PrivateKey string `json:"privateKey,omitempty"`
	// CACertificates represents the file name containing the root certificates
	// for the CA, e.g. /etc/istio/als/root-cert.pem
	// +optional
	CACertificates string `json:"caCertificates,omitempty"`
	// SNIHost represents the host name presented to the server during TLS
	// handshake, e.g. als.somedomain
	// +optional
	SNIHost string `json:"sni,omitempty"`
	// SubjectAltNames represents the list of alternative names that may be used
	// to verify the servers identity, e.g. [als.someotherdomain]
	// +optional
	SubjectAltNames []string `json:"subjectAltNames,omitempty"`
}
