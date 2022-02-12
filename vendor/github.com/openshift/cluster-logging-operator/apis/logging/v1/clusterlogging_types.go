/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/openshift/cluster-logging-operator/internal/status"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterLoggingSpec defines the desired state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator
	//
	// +kubebuilder:validation:Enum:=Managed;Unmanaged
	// +optional
	ManagementState ManagementState `json:"managementState"`

	// Specification of the Visualization component for the cluster
	//
	// +nullable
	Visualization *VisualizationSpec `json:"visualization,omitempty"`

	// Specification of the Log Storage component for the cluster
	//
	// +nullable
	LogStore *LogStoreSpec `json:"logStore,omitempty"`

	// Specification of the Collection component for the cluster
	//
	// +nullable
	Collection *CollectionSpec `json:"collection,omitempty"`

	// Specification of the Curation component for the cluster
	//
	// +nullable
	Curation *CurationSpec `json:"curation,omitempty"`

	// Specification for Forwarder component for the cluster
	//
	// +nullable
	Forwarder *ForwarderSpec `json:"forwarder,omitempty"`
}

// ClusterLoggingStatus defines the observed state of ClusterLogging
// +k8s:openapi-gen=true
type ClusterLoggingStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +optional
	Visualization VisualizationStatus `json:"visualization"`

	// +optional
	LogStore LogStoreStatus `json:"logStore"`

	// +optional
	Collection CollectionStatus `json:"collection"`

	// +optional
	Curation CurationStatus `json:"curation"`

	// +optional
	Conditions status.Conditions `json:"clusterConditions,omitempty"`
}

// This is the struct that will contain information pertinent to Log visualization (Kibana)
type VisualizationSpec struct {
	// The type of Visualization to configure
	Type VisualizationType `json:"type"`

	// Specification of the Kibana Visualization component
	KibanaSpec `json:"kibana,omitempty"`
}

type KibanaSpec struct {
	// The resource requirements for Kibana
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// Number of instances to deploy for a Kibana deployment
	Replicas *int32 `json:"replicas"`

	// Specification of the Kibana Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

type ProxySpec struct {
	// +nullable
	Resources *v1.ResourceRequirements `json:"resources"`
}

// This is the struct that will contain information pertinent to Log storage (Elasticsearch)
type LogStoreSpec struct {
	// The type of Log Storage to configure
	Type LogStoreType `json:"type"`

	// Specification of the Elasticsearch Log Store component
	ElasticsearchSpec `json:"elasticsearch,omitempty"`

	// Retention policy defines the maximum age for an index after which it should be deleted
	//
	// +nullable
	RetentionPolicy *RetentionPoliciesSpec `json:"retentionPolicy,omitempty"`
}

type RetentionPoliciesSpec struct {
	// +nullable
	App *RetentionPolicySpec `json:"application,omitempty"`

	// +nullable
	Infra *RetentionPolicySpec `json:"infra,omitempty"`

	// +nullable
	Audit *RetentionPolicySpec `json:"audit,omitempty"`
}

type RetentionPolicySpec struct {
	// +optional
	MaxAge elasticsearch.TimeUnit `json:"maxAge"`

	// How often to run a new prune-namespaces job
	// +optional
	PruneNamespacesInterval elasticsearch.TimeUnit `json:"pruneNamespacesInterval"`

	// The per namespace specification to delete documents older than a given minimum age
	// +optional
	Namespaces []elasticsearch.IndexManagementDeleteNamespaceSpec `json:"namespaceSpec,omitempty"`
}

type ElasticsearchSpec struct {
	// The resource requirements for Elasticsearch
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Number of nodes to deploy for Elasticsearch
	NodeCount int32 `json:"nodeCount"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The storage specification for Elasticsearch data nodes
	//
	// +nullable
	// +optional
	Storage elasticsearch.ElasticsearchStorageSpec `json:"storage"`

	// +optional
	RedundancyPolicy elasticsearch.RedundancyPolicyType `json:"redundancyPolicy"`

	// Specification of the Elasticsearch Proxy component
	ProxySpec `json:"proxy,omitempty"`
}

// This is the struct that will contain information pertinent to Log and event collection
type CollectionSpec struct {
	// Specification of Log Collection for the cluster
	Logs LogCollectionSpec `json:"logs,omitempty"`
}

type LogCollectionSpec struct {
	// The type of Log Collection to configure
	Type LogCollectionType `json:"type"`

	// Specification of the Fluentd Log Collection component
	FluentdSpec `json:"fluentd,omitempty"`
}

type EventCollectionSpec struct {
	Type EventCollectionType `json:"type"`
}

type FluentdSpec struct {
	// The resource requirements for Fluentd
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`
}

// This is the struct that will contain information pertinent to Log curation (Curator)
type CurationSpec struct {
	// The kind of curation to configure
	Type CurationType `json:"type"`

	// The specification of curation to configure
	CuratorSpec `json:"curator,omitempty"`
}

type CuratorSpec struct {
	// The resource requirements for Curator
	//
	// +nullable
	// +optional
	Resources *v1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  []v1.Toleration   `json:"tolerations,omitempty"`

	// The cron schedule that the Curator job is run. Defaults to "30 3 * * *"
	Schedule string `json:"schedule"`
}

// ForwarderSpec contains global tuning parameters for specific forwarder implementations.
// This field is not required for general use, it allows performance tuning by users
// familiar with the underlying forwarder technology.
// Currently supported: `fluentd`.
type ForwarderSpec struct {
	Fluentd *FluentdForwarderSpec `json:"fluentd,omitempty"`
}

// FluentdForwarderSpec represents the configuration for forwarders of type fluentd.
type FluentdForwarderSpec struct {
	Buffer *FluentdBufferSpec `json:"buffer,omitempty"`
}

const (
	// ThrowExceptionAction raises an exception when output buffer is full
	ThrowExceptionAction OverflowActionType = "throw_exception"
	// BlockAction blocks processing inputs when output buffer is full
	BlockAction OverflowActionType = "block"
	// DropOldestChunkAction drops oldest chunk to accept newly incoming chunks
	// when buffer is full
	DropOldestChunkAction OverflowActionType = "drop_oldest_chunk"
)

type OverflowActionType string

const (
	// Flush one chunk per time key if time is specified as chunk key
	FlushModeLazy FlushModeType = "lazy"
	// Flush chunks per specified time via FlushInterval
	FlushModeInterval FlushModeType = "interval"
	// Flush immediately after events appended to chunks
	FlushModeImmediate FlushModeType = "immediate"
)

type FlushModeType string

const (
	// RetryExponentialBackoff increases wait time exponentially between failures
	RetryExponentialBackoff RetryTypeType = "exponential_backoff"
	// RetryPeriodic to retry sending to output periodically on fixed intervals
	RetryPeriodic RetryTypeType = "periodic"
)

type RetryTypeType string

// FluentdSizeUnit represents fluentd's parameter type for memory sizes.
//
// For datatype pattern see:
// https://docs.fluentd.org/configuration/config-file#supported-data-types-for-values
//
// Notice: The OpenAPI validation pattern is an ECMA262 regular expression
// (See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#properties)
//
// +kubebuilder:validation:Pattern:="^([0-9]+)([kmgtKMGT]{0,1})$"
type FluentdSizeUnit string

// FluentdTimeUnit represents fluentd's parameter type for time.
//
// For data type pattern see:
// https://docs.fluentd.org/configuration/config-file#supported-data-types-for-values
//
// Notice: The OpenAPI validation pattern is an ECMA262 regular expression
// (See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#properties)
// +kubebuilder:validation:Pattern:="^([0-9]+)([smhd]{0,1})$"
type FluentdTimeUnit string

// FluentdBufferSpec represents a subset of fluentd buffer parameters to tune
// the buffer configuration for all fluentd outputs. It supports a subset of
// parameters to configure buffer and queue sizing, flush operations and retry
// flushing.
//
// For general parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#buffering-parameters
//
// For flush parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#flushing-parameters
//
// For retry parameters refer to:
// https://docs.fluentd.org/configuration/buffer-section#retries-parameters
type FluentdBufferSpec struct {
	// ChunkLimitSize represents the maximum size of each chunk. Events will be
	// written into chunks until the size of chunks become this size.
	//
	// +optional
	ChunkLimitSize FluentdSizeUnit `json:"chunkLimitSize"`

	// TotalLimitSize represents the threshold of node space allowed per fluentd
	// buffer to allocate. Once this threshold is reached, all append operations
	// will fail with error (and data will be lost).
	//
	// +optional
	TotalLimitSize FluentdSizeUnit `json:"totalLimitSize"`

	// OverflowAction represents the action for the fluentd buffer plugin to
	// execute when a buffer queue is full. (Default: block)
	//
	// +kubebuilder:validation:Enum:=throw_exception;block;drop_oldest_chunk
	// +optional
	OverflowAction OverflowActionType `json:"overflowAction"`

	// FlushThreadCount reprents the number of threads used by the fluentd buffer
	// plugin to flush/write chunks in parallel.
	//
	// +optional
	FlushThreadCount int32 `json:"flushThreadCount"`

	// FlushMode represents the mode of the flushing thread to write chunks. The mode
	// allows lazy (if `time` parameter set), per interval or immediate flushing.
	//
	// +kubebuilder:validation:Enum:=lazy;interval;immediate
	// +optional
	FlushMode FlushModeType `json:"flushMode"`

	// FlushInterval represents the time duration to wait between two consecutive flush
	// operations. Takes only effect used together with `flushMode: interval`.
	//
	// +optional
	FlushInterval FluentdTimeUnit `json:"flushInterval"`

	// RetryWait represents the time duration between two consecutive retries to flush
	// buffers for periodic retries or a constant factor of time on retries with exponential
	// backoff.
	//
	// +optional
	RetryWait FluentdTimeUnit `json:"retryWait"`

	// RetryType represents the type of retrying flush operations. Flush operations can
	// be retried either periodically or by applying exponential backoff.
	//
	// +kubebuilder:validation:Enum:=exponential_backoff;periodic
	// +optional
	RetryType RetryTypeType `json:"retryType"`

	// RetryMaxInterval represents the maxixum time interval for exponential backoff
	// between retries. Takes only effect if used together with `retryType: exponential_backoff`.
	//
	// +optional
	RetryMaxInterval FluentdTimeUnit `json:"retryMaxInterval"`

	// RetryTimeout represents the maxixum time interval to attempt retries before giving up
	// and the record is disguarded.  If unspecified, the default will be used
	//
	// +optional
	RetryTimeout FluentdTimeUnit `json:"retryTimeout"`
}

type VisualizationStatus struct {
	// +optional
	KibanaStatus []elasticsearch.KibanaStatus `json:"kibanaStatus,omitempty"`
}

type KibanaStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	Deployment string `json:"deployment"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type LogStoreStatus struct {
	// +optional
	ElasticsearchStatus []ElasticsearchStatus `json:"elasticsearchStatus,omitempty"`
}

type ElasticsearchStatus struct {
	// +optional
	ClusterName string `json:"clusterName"`
	// +optional
	NodeCount int32 `json:"nodeCount"`
	// +optional
	ReplicaSets []string `json:"replicaSets,omitempty"`
	// +optional
	Deployments []string `json:"deployments,omitempty"`
	// +optional
	StatefulSets []string `json:"statefulSets,omitempty"`
	// +optional
	ClusterHealth string `json:"clusterHealth,omitempty"`
	// +optional
	Cluster elasticsearch.ClusterHealth `json:"cluster"`
	// +optional
	Pods map[ElasticsearchRoleType]PodStateMap `json:"pods,omitempty"`
	// +optional
	ShardAllocationEnabled elasticsearch.ShardAllocationState `json:"shardAllocationEnabled"`
	// +optional
	ClusterConditions ElasticsearchClusterConditions `json:"clusterConditions,omitempty"`
	// +optional
	NodeConditions map[string]ElasticsearchClusterConditions `json:"nodeConditions,omitempty"`
}

type CollectionStatus struct {
	// +optional
	Logs LogCollectionStatus `json:"logs,omitempty"`
}

type LogCollectionStatus struct {
	// +optional
	FluentdStatus FluentdCollectorStatus `json:"fluentdStatus,omitempty"`
}

type EventCollectionStatus struct {
}

type FluentdCollectorStatus struct {
	// +optional
	DaemonSet string `json:"daemonSet,omitempty"`
	// +optional
	Nodes map[string]string `json:"nodes,omitempty"`
	// +optional
	Pods PodStateMap `json:"pods,omitempty"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type FluentdNormalizerStatus struct {
	// +optional
	Replicas int32 `json:"replicas"`
	// +optional
	ReplicaSets []string `json:"replicaSets"`
	// +optional
	Pods PodStateMap `json:"pods"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type NormalizerStatus struct {
	// +optional
	FluentdStatus []FluentdNormalizerStatus `json:"fluentdStatus,omitempty"`
}

type CurationStatus struct {
	// +optional
	CuratorStatus []CuratorStatus `json:"curatorStatus,omitempty"`
}

type CuratorStatus struct {
	// +optional
	CronJob string `json:"cronJobs"`
	// +optional
	Schedule string `json:"schedules"`
	// +optional
	Suspended bool `json:"suspended"`
	// +optional
	Conditions map[string]ClusterConditions `json:"clusterCondition,omitempty"`
}

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type LogStoreType string

const (
	LogStoreTypeElasticsearch LogStoreType = "elasticsearch"
)

type ElasticsearchRoleType string

const (
	ElasticsearchRoleTypeClient ElasticsearchRoleType = "client"
	ElasticsearchRoleTypeData   ElasticsearchRoleType = "data"
	ElasticsearchRoleTypeMaster ElasticsearchRoleType = "master"
)

type VisualizationType string

const (
	VisualizationTypeKibana VisualizationType = "kibana"
)

type CurationType string

const (
	CurationTypeCurator CurationType = "curator"
)

type LogCollectionType string

const (
	LogCollectionTypeFluentd LogCollectionType = "fluentd"
	LogCollectionTypeVector  LogCollectionType = "vector"
)

type EventCollectionType string

type NormalizerType string

type ManagementState string

const (
	// Managed means that the operator is actively managing its resources and trying to keep the component active.
	// It will only upgrade the component if it is safe to do so
	ManagementStateManaged ManagementState = "Managed"
	// Unmanaged means that the operator will not take any action related to the component
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

const (
	IncorrectCRName     ConditionType = "IncorrectCRName"
	ContainerWaiting    ConditionType = "ContainerWaiting"
	ContainerTerminated ConditionType = "ContainerTerminated"
	Unschedulable       ConditionType = "Unschedulable"
	NodeStorage         ConditionType = "NodeStorage"
	CollectorDeadEnd    ConditionType = "CollectorDeadEnd"
)

// `operator-sdk generate crds` does not allow map-of-slice, must use a named type.
type ClusterConditions []Condition
type ElasticsearchClusterConditions []elasticsearch.ClusterCondition

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=logging,shortName=cl
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// ClusterLogging is the Schema for the clusterloggings API
// +operator-sdk:csv:customresourcedefinitions:displayName="ClusterLogging",resources={{Pod,v1},{Deployment,v1},{ReplicaSet,v1},{ConfigMap,v1},{Service,v1},{Route,v1},{CronJob,v1beta1},{Role,v1},{RoleBinding,v1},{ServiceAccount,v1},{ServiceMonitor,v1},{persistentvolumeclaims,v1}}
type ClusterLogging struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLoggingSpec   `json:"spec,omitempty"`
	Status ClusterLoggingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterLoggingList contains a list of ClusterLogging
type ClusterLoggingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLogging `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLogging{}, &ClusterLoggingList{})
}
