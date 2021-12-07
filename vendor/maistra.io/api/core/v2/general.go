package v2

// GeneralConfig for control plane
type GeneralConfig struct {
	// Logging represents the logging configuration for the control plane components
	// XXX: Should this be separate from Proxy.Logging?
	// +optional
	Logging *LoggingConfig `json:"logging,omitempty"`

	// ValidationMessages configures the control plane to add validationMessages
	// to the status fields of istio.io resources.  This can be usefule for
	// detecting configuration errors in resources.
	// .Values.galley.enableAnalysis (<v2.0)
	// .Values.global.istiod.enableAnalysis (>=v2.0)
	ValidationMessages *bool `json:"validationMessages,omitempty"`
}
