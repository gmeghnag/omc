package v2

// SecurityConfig specifies security aspects of the control plane.
type SecurityConfig struct {
	// Trust configures trust aspects associated with mutual TLS clients.
	// +optional
	Trust *TrustConfig `json:"trust,omitempty"`
	// CertificateAuthority configures the certificate authority used by the
	// control plane to create and sign client certs and server keys.
	// +optional
	CertificateAuthority *CertificateAuthorityConfig `json:"certificateAuthority,omitempty"`
	// Identity configures the types of user tokens used by clients.
	// +optional
	Identity *IdentityConfig `json:"identity,omitempty"`
	// ControlPlane configures mutual TLS for control plane communication.
	// +optional
	ControlPlane *ControlPlaneSecurityConfig `json:"controlPlane,omitempty"`
	// DataPlane configures mutual TLS for data plane communication.
	// +optional
	DataPlane *DataPlaneSecurityConfig `json:"dataPlane,omitempty"`
}

// TrustConfig configures trust aspects associated with mutual TLS clients
type TrustConfig struct {
	// Domain specifies the trust domain to be used by the mesh.
	//.Values.global.trustDomain, maps to trustDomain
	// The trust domain corresponds to the trust root of a system.
	// Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
	// +optional
	Domain string `json:"domain,omitempty"`
	// AdditionalDomains are additional SPIFFE trust domains that are accepted as trusted.
	// .Values.global.trustDomainAliases, maps to trustDomainAliases
	//  Any service with the identity "td1/ns/foo/sa/a-service-account", "td2/ns/foo/sa/a-service-account",
	//  or "td3/ns/foo/sa/a-service-account" will be treated the same in the Istio mesh.
	// +optional
	AdditionalDomains []string `json:"additionalDomains,omitempty"`
}

// CertificateAuthorityConfig configures the certificate authority implementation
// used by the control plane.
type CertificateAuthorityConfig struct {
	// Type is the certificate authority to use.
	Type CertificateAuthorityType `json:"type,omitempty"`
	// Istiod is the configuration for Istio's internal certificate authority implementation.
	// each of these produces a CAEndpoint, i.e. CA_ADDR
	// +optional
	Istiod *IstiodCertificateAuthorityConfig `json:"istiod,omitempty"`
	// Custom is the configuration for a custom certificate authority.
	// +optional
	Custom *CustomCertificateAuthorityConfig `json:"custom,omitempty"`
}

// CertificateAuthorityType represents the type of CertificateAuthority implementation.
type CertificateAuthorityType string

const (
	// CertificateAuthorityTypeIstiod represents Istio's internal certificate authority implementation
	CertificateAuthorityTypeIstiod CertificateAuthorityType = "Istiod"
	// CertificateAuthorityTypeCustom represents a custom certificate authority implementation
	CertificateAuthorityTypeCustom CertificateAuthorityType = "Custom"
)

// IstiodCertificateAuthorityConfig is the configuration for Istio's internal
// certificate authority implementation.
type IstiodCertificateAuthorityConfig struct {
	// Type of certificate signer to use.
	Type IstioCertificateSignerType `json:"type,omitempty"`
	// SelfSigned configures istiod to generate and use a self-signed certificate for the root.
	// +optional
	SelfSigned *IstioSelfSignedCertificateSignerConfig `json:"selfSigned,omitempty"`
	// PrivateKey configures istiod to use a user specified private key/cert when signing certificates.
	// +optional
	PrivateKey *IstioPrivateKeyCertificateSignerConfig `json:"privateKey,omitempty"`
	// WorkloadCertTTLDefault is the default TTL for generated workload
	// certificates.  Used if not specified in CSR (<= 0)
	// env DEFAULT_WORKLOAD_CERT_TTL, 1.6
	// --workload-cert-ttl, citadel, pre-1.6
	// defaults to 24 hours
	// +optional
	WorkloadCertTTLDefault string `json:"workloadCertTTLDefault,omitempty"`
	// WorkloadCertTTLMax is the maximum TTL for generated workload certificates.
	// env MAX_WORKLOAD_CERT_TTL
	// --max-workload-cert-ttl, citadel, pre-1.6
	// defaults to 90 days
	// +optional
	WorkloadCertTTLMax string `json:"workloadCertTTLMax,omitempty"`
}

// IstioCertificateSignerType represents the certificate signer implementation used by istiod.
type IstioCertificateSignerType string

const (
	// IstioCertificateSignerTypePrivateKey is the signer type used when signing with a user specified private key.
	IstioCertificateSignerTypePrivateKey IstioCertificateSignerType = "PrivateKey"
	// IstioCertificateSignerTypeSelfSigned is the signer type used when signing with a generated, self-signed certificate.
	IstioCertificateSignerTypeSelfSigned IstioCertificateSignerType = "SelfSigned"
)

// IstioSelfSignedCertificateSignerConfig is the configuration for using a
// self-signed root certificate.
type IstioSelfSignedCertificateSignerConfig struct {
	// TTL for self-signed root certificate
	// env CITADEL_SELF_SIGNED_CA_CERT_TTL
	// default is 10 years
	// +optional
	TTL string `json:"ttl,omitempty"`
	// GracePeriod percentile for self-signed cert
	// env CITADEL_SELF_SIGNED_ROOT_CERT_GRACE_PERIOD_PERCENTILE
	// default is 20%
	// +optional
	GracePeriod string `json:"gracePeriod,omitempty"`
	// CheckPeriod is the interval with which certificate is checked for rotation
	// env CITADEL_SELF_SIGNED_ROOT_CERT_CHECK_INTERVAL
	// default is 1 hour, zero or negative value disables cert rotation
	// +optional
	CheckPeriod string `json:"checkPeriod,omitempty"`
	// EnableJitter to use jitter for cert rotation
	// env CITADEL_ENABLE_JITTER_FOR_ROOT_CERT_ROTATOR
	// defaults to true
	// +optional
	EnableJitter *bool `json:"enableJitter,omitempty"`
	// Org is the Org value in the certificate.
	// XXX: currently uses TrustDomain.  I don't think this is configurable.
	// +optional
	//Org string `json:"org,omitempty"`
}

// IstioPrivateKeyCertificateSignerConfig is the configuration when using a user
// supplied private key/cert for signing.
// XXX: nothing in here is currently configurable, except RootCADir
type IstioPrivateKeyCertificateSignerConfig struct {
	// hard coded to use a secret named cacerts
	// +optional
	//EncryptionSecret string `json:"encryptionSecret,omitempty"`
	// ROOT_CA_DIR, defaults to /etc/cacerts
	// Mount directory for encryption secret
	// XXX: currently, not configurable in the charts
	// +optional
	RootCADir string `json:"rootCADir,omitempty"`
	// hard coded to ca-key.pem
	// +optional
	//SigningKeyFile string `json:"signingKeyFile,omitempty"`
	// hard coded to ca-cert.pem
	// +optional
	//SigningCertFile string `json:"signingCertFile,omitempty"`
	// hard coded to root-cert.pem
	// +optional
	//RootCertFile string `json:"rootCertFile,omitempty"`
	// hard coded to cert-chain.pem
	// +optional
	//CertChainFile string `json:"certChainFile,omitempty"`
}

// CustomCertificateAuthorityConfig is the configuration for a custom
// certificate authority.
type CustomCertificateAuthorityConfig struct {
	// Address is the grpc address for an Istio compatible certificate authority endpoint.
	// .Values.global.caAddress
	// XXX: assumption is this is a grpc endpoint that provides methods like istio.v1.auth.IstioCertificateService/CreateCertificate
	Address string `json:"address,omitempty"`
}

// IdentityConfig configures the types of user tokens used by clients
type IdentityConfig struct {
	// Type is the type of identity tokens being used.
	// .Values.global.jwtPolicy
	Type IdentityConfigType `json:"type,omitempty"`
	// ThirdParty configures istiod to use a third-party token provider for
	// identifying users. (basically uses a custom audience, e.g. istio-ca)
	// XXX: this is only supported on OCP 4.4+
	// +optional
	ThirdParty *ThirdPartyIdentityConfig `json:"thirdParty,omitempty"`
}

// IdentityConfigType represents the identity implementation being used.
type IdentityConfigType string

const (
	// IdentityConfigTypeKubernetes specifies Kubernetes as the token provider.
	IdentityConfigTypeKubernetes IdentityConfigType = "Kubernetes" // first-party-jwt
	// IdentityConfigTypeThirdParty specifies a third-party token provider.
	IdentityConfigTypeThirdParty IdentityConfigType = "ThirdParty" // third-party-jwt
)

// ThirdPartyIdentityConfig configures a third-party token provider for use with
// istiod.
type ThirdPartyIdentityConfig struct {
	// TokenPath is the path to the token used to identify the workload.
	// default /var/run/secrets/tokens/istio-token
	// XXX: projects service account token with specified audience (istio-ca)
	// XXX: not configurable
	// +optional
	//TokenPath string `json:"tokenPath,omitempty"`

	// Issuer is the URL of the issuer.
	// env TOKEN_ISSUER, defaults to iss in specified token
	// only supported in 1.6+
	// +optional
	Issuer string `json:"issuer,omitempty"`
	// Audience is the audience for whom the token is intended.
	// env AUDIENCE
	// .Values.global.sds.token.aud, defaults to istio-ca
	// +optional
	Audience string `json:"audience,omitempty"`
}

// ControlPlaneSecurityConfig is the mutual TLS configuration specific to the
// control plane.
type ControlPlaneSecurityConfig struct {
	// Enable mutual TLS for the control plane components.
	// .Values.global.controlPlaneSecurityEnabled
	// +optional
	MTLS *bool `json:"mtls,omitempty"`
	// CertProvider is the certificate authority used to generate the serving
	// certificates for the control plane components.
	// .Values.global.pilotCertProvider
	// Provider used to generate serving certs for istiod (pilot)
	// +optional
	CertProvider ControlPlaneCertProviderType `json:"certProvider,omitempty"`

	// TLS configures aspects of TLS listeners created by control plane components.
	// +optional
	TLS *ControlPlaneTLSConfig `json:"tls,omitempty"`
}

// DataPlaneSecurityConfig is the mutual TLS configuration specific to the
// control plane.
type DataPlaneSecurityConfig struct {
	// Enable mutual TLS by default.
	// .Values.global.mtls.enabled
	MTLS *bool `json:"mtls,omitempty"`
	// Auto configures the mesh to automatically detect whether or not mutual
	// TLS is required for a specific connection.
	// .Values.global.mtls.auto
	// +optional
	AutoMTLS *bool `json:"automtls,omitempty"`
}

// ControlPlaneCertProviderType represents the provider used to generate serving
// certificates for the control plane.
type ControlPlaneCertProviderType string

const (
	// ControlPlaneCertProviderTypeIstiod identifies istiod as the provider generating the serving certifications.
	ControlPlaneCertProviderTypeIstiod ControlPlaneCertProviderType = "Istiod"
	// ControlPlaneCertProviderTypeKubernetes identifies Kubernetes as the provider generating the serving certificates.
	ControlPlaneCertProviderTypeKubernetes ControlPlaneCertProviderType = "Kubernetes"
	// ControlPlaneCertProviderTypeCustom identifies a custom provider has generated the serving certificates.
	// XXX: Not quite sure what this means. Presumably, the key and cert chain have been mounted specially
	ControlPlaneCertProviderTypeCustom ControlPlaneCertProviderType = "Custom"
)

// ControlPlaneTLSConfig configures settings on TLS listeners created by
// control plane components, e.g. webhooks, grpc (if mtls is enabled), etc.
type ControlPlaneTLSConfig struct {
	// CipherSuites configures the cipher suites that are available for use by
	// TLS listeners.
	// .Values.global.tls.cipherSuites
	// +optional
	CipherSuites []string `json:"cipherSuites,omitempty"`
	// ECDHCurves configures the ECDH curves that are available for use by
	// TLS listeners.
	// .Values.global.tls.ecdhCurves
	// +optional
	ECDHCurves []string `json:"ecdhCurves,omitempty"`
	// MinProtocolVersion the minimum TLS version that should be supported by
	// the listeners.
	// .Values.global.tls.minProtocolVersion
	// +optional
	MinProtocolVersion string `json:"minProtocolVersion,omitempty"`
	// MaxProtocolVersion the maximum TLS version that should be supported by
	// the listeners.
	// .Values.global.tls.maxProtocolVersion
	// +optional
	MaxProtocolVersion string `json:"maxProtocolVersion,omitempty"`
}
