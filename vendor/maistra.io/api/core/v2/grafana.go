package v2

// GrafanaAddonConfig configures a grafana instance for use with the mesh. Only
// one of install or address may be specified
type GrafanaAddonConfig struct {
	Enablement `json:",inline"`
	// Install a new grafana instance and manage with control plane
	// +optional
	Install *GrafanaInstallConfig `json:"install,omitempty"`
	// Address is the address of an existing grafana installation
	// implies .Values.kiali.dashboard.grafanaURL
	// +optional
	Address *string `json:"address,omitempty"`
}

// GrafanaInstallConfig is used to configure a new installation of grafana.
type GrafanaInstallConfig struct {
	// SelfManaged, true if the entire install should be managed by Maistra, false if using grafana CR (not supported)
	// +optional
	SelfManaged bool `json:"selfManaged,omitempty"`
	// Config configures the behavior of the grafana installation
	// +optional
	Config *GrafanaConfig `json:"config,omitempty"`
	// Service configures the k8s Service associated with the grafana installation
	// XXX: grafana service config does not follow other addon components' structure
	// +optional
	Service *ComponentServiceConfig `json:"service,omitempty"`
	// Persistence configures a PersistentVolume associated with the grafana installation
	// .Values.grafana.persist, true if not null
	// XXX: capacity is not supported in the charts, hard coded to 5Gi
	// +optional
	Persistence *ComponentPersistenceConfig `json:"persistence,omitempty"`
	// Security is used to secure the grafana service.
	// .Values.grafana.security.enabled, true if not null
	// XXX: unused for maistra, as we use oauth-proxy
	// +optional
	Security *GrafanaSecurityConfig `json:"security,omitempty"`
}

// GrafanaConfig configures the behavior of the grafana installation
type GrafanaConfig struct {
	// Env allows specification of various grafana environment variables to be
	// configured on the grafana container.
	// .Values.grafana.env
	// XXX: This is pretty cheesy...
	// +optional
	Env map[string]string `json:"env,omitempty"`
	// EnvSecrets allows specification of secret fields into grafana environment
	// variables to be configured on the grafana container
	// .Values.grafana.envSecrets
	// XXX: This is pretty cheesy...
	// +optional
	EnvSecrets map[string]string `json:"envSecrets,omitempty"`
}

// GrafanaSecurityConfig is used to secure access to grafana
type GrafanaSecurityConfig struct {
	Enablement `json:",inline"`
	// SecretName is the name of a secret containing the username/password that
	// should be used to access grafana.
	// +optional
	SecretName string `json:"secretName,omitempty"`
	// UsernameKey is the name of the key within the secret identifying the username.
	// +optional
	UsernameKey string `json:"usernameKey,omitempty"`
	// PassphraseKey is the name of the key within the secret identifying the password.
	// +optional
	PassphraseKey string `json:"passphraseKey,omitempty"`
}
