package v1

// NOTE: The Enum validation on OutputSpec.Type must be updated if the list of
// known types changes.

// Output type constants, must match JSON tags of OutputTypeSpec fields.
const (
	OutputTypeCloudwatch     = "cloudwatch"
	OutputTypeElasticsearch  = "elasticsearch"
	OutputTypeFluentdForward = "fluentdForward"
	OutputTypeSyslog         = "syslog"
	OutputTypeKafka          = "kafka"
	OutputTypeLoki           = "loki"
)

// OutputTypeSpec is a union of optional additional configuration specific to an
// output type. The fields of this struct define the set of known output types.
type OutputTypeSpec struct {
	// +optional
	Syslog *Syslog `json:"syslog,omitempty"`
	// +optional
	FluentdForward *FluentdForward `json:"fluentdForward,omitempty"`
	// +optional
	Elasticsearch *Elasticsearch `json:"elasticsearch,omitempty"`
	// +optional
	Kafka *Kafka `json:"kafka,omitempty"`
	// +optional
	Cloudwatch *Cloudwatch `json:"cloudwatch,omitempty"`
	// +optional
	Loki *Loki `json:"loki,omitempty"`
}

// Cloudwatch provides configuration for the output type `cloudwatch`
//
// Note: the cloudwatch output recognizes the following additional keys in the Secret:
//
//	`aws_secret_access_key`: AWS secret access key.
// 	`aws_access_key_id`:AWS secret access key ID.
//
type Cloudwatch struct {
	// +required
	Region string `json:"region,omitempty"`

	//GroupBy defines the strategy for grouping logstreams
	// +required
	//+kubebuilder:validation:Enum:=logType;namespaceName;namespaceUUID
	GroupBy LogGroupByType `json:"groupBy,omitempty"`

	//GroupPrefix Add this prefix to all group names.
	//  Useful to avoid group name clashes if an AWS account is used for multiple clusters and
	//  used verbatim (e.g. "" means no prefix)
	//  The default prefix is cluster-name/log-type
	// +optional
	GroupPrefix *string `json:"groupPrefix,omitempty"`
}

// LogGroupByType defines a fixed strategy type
type LogGroupByType string

const (
	//LogGroupByLogType is the strategy to group logs by source(e.g. app, infra)
	LogGroupByLogType LogGroupByType = "logType"

	// LogGroupByNamespaceName is the strategy to use for grouping logs by namespace. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceName LogGroupByType = "namespaceName"

	// LogGroupByNamespaceUUID  is the strategy to use for grouping logs by namespace UUID. Infrastructure and
	// audit logs are always grouped by "logType"
	LogGroupByNamespaceUUID LogGroupByType = "namespaceUUID"
)

// Syslog provides optional extra properties for output type `syslog`
type Syslog struct {
	// Severity to set on outgoing syslog records.
	//
	// Severity values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1
	// The value can be a decimal integer or one of these case-insensitive keywords:
	//
	//     Emergency Alert Critical Error Warning Notice Informational Debug
	//
	// +optional
	Severity string `json:"severity,omitempty"`

	// Facility to set on outgoing syslog records.
	//
	// Facility values are defined in https://tools.ietf.org/html/rfc5424#section-6.2.1.
	// The value can be a decimal integer. Facility keywords are not standardized,
	// this API recognizes at least the following case-insensitive keywords
	// (defined by https://en.wikipedia.org/wiki/Syslog#Facility_Levels):
	//
	//     kernel user mail daemon auth syslog lpr news
	//     uucp cron authpriv ftp ntp security console solaris-cron
	//     local0 local1 local2 local3 local4 local5 local6 local7
	//
	// +optional
	Facility string `json:"facility,omitempty"`

	// TrimPrefix is a prefix to trim from the tag.
	//
	// +optional
	TrimPrefix string `json:"trimPrefix,omitempty"`

	// Tag specifies a record field to use as tag.
	//
	// +optional
	Tag string `json:"tag,omitempty"`

	// PayloadKey specifies record field to use as payload.
	//
	// +optional
	PayloadKey string `json:"payloadKey,omitempty"`

	// AddLogSource adds log's source information to the log message
	// If the logs are collected from a process; namespace_name, pod_name, container_name is added to the log
	// In addition, it picks the originating process name and id(known as the `pid`) from the record
	// and injects them into the header field."
	//
	// +optional
	AddLogSource bool `json:"addLogSource,omitempty"`

	// Rfc specifies the rfc to be used for sending syslog
	//
	// Rfc values can be one of:
	//  - RFC3164 (https://tools.ietf.org/html/rfc3164)
	//  - RFC5424 (https://tools.ietf.org/html/rfc5424)
	//
	// If unspecified, RFC5424 will be assumed.
	//
	// +kubebuilder:validation:Enum:=RFC3164;RFC5424
	// +kubebuilder:default:=RFC5424
	// +optional
	RFC string `json:"rfc,omitempty"`

	// AppName is APP-NAME part of the syslog-msg header
	//
	// AppName needs to be specified if using rfc5424
	//
	// +optional
	AppName string `json:"appName,omitempty"`

	// ProcID is PROCID part of the syslog-msg header
	//
	// ProcID needs to be specified if using rfc5424
	//
	// +optional
	ProcID string `json:"procID,omitempty"`

	// MsgID is MSGID part of the syslog-msg header
	//
	// MsgID needs to be specified if using rfc5424
	//
	// +optional
	MsgID string `json:"msgID,omitempty"`
}

// Kafka provides optional extra properties for `type: kafka`
type Kafka struct {
	// Topic specifies the target topic to send logs to.
	//
	// +optional
	Topic string `json:"topic,omitempty"`

	// Brokers specifies the list of brokers
	// to register in addition to the main output URL
	// on initial connect to enhance reliability.
	//
	// +optional
	Brokers []string `json:"brokers,omitempty"`
}

// FluentdForward does not provide additional fields, but note that
// the fluentforward output allows this additional keys in the Secret:
//
//   `shared_key`: (string) Key to enable fluent-forward shared-key authentication.
type FluentdForward struct{}

type Elasticsearch struct {
	// StructuredTypeKey specifies the metadata key to be used as name of elasticsearch index
	// It takes precedence over StructuredTypeName
	//
	// +optional
	StructuredTypeKey string `json:"structuredTypeKey,omitempty"`

	// StructuredTypeName specifies the name of elasticsearch schema
	//
	// +optional
	StructuredTypeName string `json:"structuredTypeName,omitempty"`
}

// Loki provides optional extra properties for `type: loki`
type Loki struct {
	// TenantKey is a meta-data key field to use as the TenantID,
	// For example: 'TenantKey: kubernetes.namespace_name` will use the kubernetes
	// namespace as the tenant ID.
	//
	// +optional
	TenantKey string `json:"tenantKey,omitempty"`

	// LabelKeys is a list of meta-data field keys to replace the default Loki labels.
	//
	// Loki label names must match the regular expression "[a-zA-Z_:][a-zA-Z0-9_:]*".
	// Illegal characters in meta-data keys are replaced with "_" to form the label name.
	// For example meta-data key "kubernetes.labels.foo" becomes Loki label "kubernetes_labels_foo".
	//
	// If LabelKeys is not set, the default keys are `[log_type, kubernetes.namespace_name, kubernetes.pod_name, kubernetes_host]`
	// These keys are translated to Loki labels by replacing '.' with '_' as: `log_type`, `kubernetes_namespace_name`, `kubernetes_pod_name`, `kubernetes_host`
	// Note that not all logs will include all of these keys: audit logs and infrastructure journal logs do not have namespace or pod name.
	//
	// Note: the set of labels should be small, Loki imposes limits on the size and number of labels allowed.
	// See https://grafana.com/docs/loki/latest/configuration/#limits_config for more.
	// You can still query based on any log record field using query filters.
	//
	// +optional
	LabelKeys []string `json:"labelKeys,omitempty"`
}
