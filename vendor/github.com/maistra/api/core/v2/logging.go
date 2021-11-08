package v2

// LoggingConfig for control plane components
type LoggingConfig struct {
	// ComponentLevels configures log level for specific envoy components
	// .Values.global.proxy.componentLogLevel, overridden by sidecar.istio.io/componentLogLevel
	// map of <component>:<level>
	// +optional
	ComponentLevels ComponentLogLevels `json:"componentLevels,omitempty"`
	// LogAsJSON enables JSON logging
	// .Values.global.logAsJson
	// +optional
	LogAsJSON *bool `json:"logAsJSON,omitempty"`
}

// ProxyLoggingConfig configures logging for a component
type ProxyLoggingConfig struct {
	// Level the log level
	// .Values.global.proxy.logLevel, overridden by sidecar.istio.io/logLevel
	// +optional
	Level LogLevel `json:"level,omitempty"`
	// ComponentLevels configures log level for specific envoy components
	// .Values.global.proxy.componentLogLevel, overridden by sidecar.istio.io/componentLogLevel
	// map of <component>:<level>
	// +optional
	ComponentLevels ComponentLogLevels `json:"componentLevels,omitempty"`
}

// ComponentLogLevels represent various logging levels, e.g. trace, debug, etc.
type ComponentLogLevels map[EnvoyComponent]LogLevel

// LogLevel represents the logging level
type LogLevel string

const (
	// LogLevelTrace trace logging level
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug debug logging level
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo info logging level
	LogLevelInfo LogLevel = "info"
	// LogLevelWarning warning logging level
	LogLevelWarning LogLevel = "warn"
	// LogLevelWarningProxy proxy warning logging level
	LogLevelWarningProxy LogLevel = "warning"
	// LogLevelError error logging level
	LogLevelError LogLevel = "error"
	// LogLevelCritical critical logging level
	LogLevelCritical LogLevel = "critical"
	// LogLevelOff disable logging
	LogLevelOff LogLevel = "off"
)

// EnvoyComponent represents an envoy component to configure logging
type EnvoyComponent string

// not a comprehensive list
const (
	EnvoyComponentAdmin         EnvoyComponent = "admin"
	EnvoyComponentAssert        EnvoyComponent = "assert"
	EnvoyComponentBacktrace     EnvoyComponent = "backtrace"
	EnvoyComponentClient        EnvoyComponent = "client"
	EnvoyComponentConfig        EnvoyComponent = "config"
	EnvoyComponentConnection    EnvoyComponent = "connection"
	EnvoyComponentConnHandler   EnvoyComponent = "conn_handler"
	EnvoyComponentFile          EnvoyComponent = "file"
	EnvoyComponentFilter        EnvoyComponent = "filter"
	EnvoyComponentForwardProxy  EnvoyComponent = "forward_proxy"
	EnvoyComponentGRPC          EnvoyComponent = "grpc"
	EnvoyComponentHealth        EnvoyComponent = "hc"
	EnvoyComponentHealthChecker EnvoyComponent = "health_checker"
	EnvoyComponentHTTP          EnvoyComponent = "http"
	EnvoyComponentHTTP2         EnvoyComponent = "http2"
	EnvoyComponentInit          EnvoyComponent = "init"
	EnvoyComponentIO            EnvoyComponent = "io"
	EnvoyComponentJWT           EnvoyComponent = "jwt"
	EnvoyComponentLua           EnvoyComponent = "lua"
	EnvoyComponentMain          EnvoyComponent = "main"
	EnvoyComponentMisc          EnvoyComponent = "misc"
	EnvoyComponentQuic          EnvoyComponent = "quic"
	EnvoyComponentPool          EnvoyComponent = "pool"
	EnvoyComponentRBAC          EnvoyComponent = "rbac"
	EnvoyComponentRouter        EnvoyComponent = "router"
	EnvoyComponentRuntime       EnvoyComponent = "runtime"
	EnvoyComponentStats         EnvoyComponent = "stats"
	EnvoyComponentSecret        EnvoyComponent = "secret"
	EnvoyComponentTap           EnvoyComponent = "tap"
	EnvoyComponentTesting       EnvoyComponent = "testing"
	EnvoyComponentTracing       EnvoyComponent = "tracing"
	EnvoyComponentUpstream      EnvoyComponent = "upstream"
	EnvoyComponentUDP           EnvoyComponent = "udp"
	EnvoyComponentWASM          EnvoyComponent = "wasm"
)
