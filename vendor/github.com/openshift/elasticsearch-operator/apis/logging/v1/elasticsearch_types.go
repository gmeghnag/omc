package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServiceAccountName string = "elasticsearch"
	ConfigMapName      string = "elasticsearch"
	SecretName         string = "elasticsearch"
)

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories=logging;tracing,shortName=es
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Management State",JSONPath=".spec.managementState",type=string
// +kubebuilder:printcolumn:name="Health",JSONPath=".status.cluster.status",type=string
// +kubebuilder:printcolumn:name="Nodes",JSONPath=".status.cluster.numNodes",type=integer
// +kubebuilder:printcolumn:name="Data Nodes",JSONPath=".status.cluster.numDataNodes",type=integer
// +kubebuilder:printcolumn:name="Shard Allocation",JSONPath=".status.shardAllocationEnabled",type=string
// +kubebuilder:printcolumn:name="Index Management",JSONPath=".status.indexManagement.State",type=string
//
// An Elasticsearch cluster instance
// +operator-sdk:csv:customresourcedefinitions:displayName="Elasticsearch",resources={{Pod,v1},{Deployment,v1},{StatefulSet,v1},{ReplicaSet,v1},{ConfigMap,v1},{Service,v1},{Route,v1},{CronJob,v1beta1},{PrometheusRule,v1},{Role,v1},{RoleBinding,v1},{ServiceAccount,v1},{ServiceMonitor,v1},{persistentvolumeclaims,v1}}
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Elasticsearch cluster
	Spec   ElasticsearchSpec   `json:"spec,omitempty"`
	Status ElasticsearchStatus `json:"status,omitempty"`
}

// AddOwnerRefTo appends the Elasticsearch object as an OwnerReference to the passed object
func (es *Elasticsearch) AddOwnerRefTo(o metav1.Object) {
	ref := es.GetOwnerRef()
	o.SetOwnerReferences(append(o.GetOwnerReferences(), ref))
}

func (es *Elasticsearch) GetOwnerRef() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: GroupVersion.String(),
		Kind:       "Elasticsearch",
		Name:       es.Name,
		UID:        es.UID,
		Controller: &trueVar,
	}
}

// +kubebuilder:object:root=true
//
// ElasticsearchList contains a list of Elasticsearch
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elasticsearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Elasticsearch{}, &ElasticsearchList{})
}

// ElasticsearchSpec defines the desired state of Elasticsearch
type ElasticsearchSpec struct {

	// ManagementState indicates whether and how the operator should manage the component.
	// Indicator if the resource is 'Managed' or 'Unmanaged' by the operator.
	ManagementState ManagementState `json:"managementState"`

	// The policy towards data redundancy to specify the number of redundant primary shards
	RedundancyPolicy RedundancyPolicyType `json:"redundancyPolicy"`

	// Specification of the different Elasticsearch nodes
	//
	// +optional
	Nodes []ElasticsearchNode `json:"nodes"`

	// Default specification applied to all Elasticsearch nodes
	//
	// +optional
	Spec ElasticsearchNodeSpec `json:"nodeSpec"`

	// Management spec for indicies
	//
	// +nullable
	// +optional
	IndexManagement *IndexManagementSpec `json:"indexManagement"`
}

// ElasticsearchStatus defines the observed state of Elasticsearch
// +k8s:openapi-gen=true
type ElasticsearchStatus struct {
	// +nullable
	// +optional
	Nodes []ElasticsearchNodeStatus `json:"nodes,omitempty"`
	// +optional
	ClusterHealth string `json:"clusterHealth,omitempty"`
	// +optional
	Cluster ClusterHealth `json:"cluster,omitempty"`
	// +optional
	ShardAllocationEnabled ShardAllocationState `json:"shardAllocationEnabled,omitempty"`
	// +nullable
	// +optional
	Pods map[ElasticsearchNodeRole]PodStateMap `json:"pods,omitempty"`
	// +optional
	Conditions ClusterConditions `json:"conditions,omitempty"`
	// +optional
	IndexManagementStatus *IndexManagementStatus `json:"indexManagement,omitempty"`
}

type ClusterHealth struct {
	// The current Status of the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:io.kubernetes.phase"
	Status string `json:"status"`
	// The number of Nodes for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Number of Nodes",xDescriptors="urn:alm:descriptor:text"
	NumNodes int32 `json:"numNodes"`
	// The number of Data Nodes for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Number of Data Nodes",xDescriptors="urn:alm:descriptor:text"
	NumDataNodes int32 `json:"numDataNodes"`

	// The number of Active Primary Shards for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text"
	ActivePrimaryShards int32 `json:"activePrimaryShards"`

	// The number of Active Shards for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text"
	ActiveShards int32 `json:"activeShards"`
	// The number of Relocating Shards for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text"
	RelocatingShards int32 `json:"relocatingShards"`
	// The number of Initializing Shards for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text"
	InitializingShards int32 `json:"initializingShards"`
	// The number of Unassigned Shards for the Elasticsearch Cluster
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors="urn:alm:descriptor:text"
	UnassignedShards int32 `json:"unassignedShards"`
	PendingTasks     int32 `json:"pendingTasks"`
}

// ElasticsearchNode struct represents individual node in Elasticsearch cluster
type ElasticsearchNode struct {
	// The specific Elasticsearch cluster roles the node should perform
	//
	// +optional
	Roles []ElasticsearchNodeRole `json:"roles"`

	// Number of nodes to deploy
	//
	// +optional
	NodeCount int32 `json:"nodeCount"`

	// The resource requirements for the Elasticsearch node
	//
	// +nullable
	// +optional
	Resources corev1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// The type of backing storage that should be used for the node
	//
	// +optional
	Storage ElasticsearchStorageSpec `json:"storage,omitempty"`

	// GenUUID will be populated by the operator if not provided
	//
	// +nullable
	GenUUID *string `json:"genUUID,omitempty"`

	// The resource requirements for the Elasticsearch proxy
	ProxyResources corev1.ResourceRequirements `json:"proxyResources,omitempty"`
}

// ElasticsearchNodeSpec represents configuration of an individual Elasticsearch node
type ElasticsearchNodeSpec struct {
	// The image to use for the Elasticsearch nodes
	//
	// +nullable
	// +optional
	Image string `json:"image,omitempty"`

	// The resource requirements for the Elasticsearch nodes
	//
	// +nullable
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resource Requirements",xDescriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources corev1.ResourceRequirements `json:"resources"`

	// Define which Nodes the Pods are scheduled on.
	//
	// +nullable
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// The resource requirements for the Elasticsearch proxy
	//
	// +nullable
	// +optional
	ProxyResources corev1.ResourceRequirements `json:"proxyResources,omitempty"`
}

type ElasticsearchStorageSpec struct {
	// The name of the storage class to use with creating the node's PVC.
	// More info: https://kubernetes.io/docs/concepts/storage/storage-classes/
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// The max storage capacity for the node to provision.
	Size *resource.Quantity `json:"size,omitempty"`
}

// ElasticsearchNodeStatus represents the status of individual Elasticsearch node
type ElasticsearchNodeStatus struct {
	// +optional
	DeploymentName string `json:"deploymentName,omitempty"`
	// +optional
	StatefulSetName string `json:"statefulSetName,omitempty"`
	// +optional
	Status string `json:"status,omitempty"`
	// +optional
	UpgradeStatus ElasticsearchNodeUpgradeStatus `json:"upgradeStatus,omitempty"`
	// +optional
	Roles []ElasticsearchNodeRole `json:"roles,omitempty"`
	// +optional
	Conditions ClusterConditions `json:"conditions,omitempty"`
}

type ElasticsearchNodeUpgradeStatus struct {
	ScheduledForUpgrade      corev1.ConditionStatus    `json:"scheduledUpgrade,omitempty"`
	ScheduledForRedeploy     corev1.ConditionStatus    `json:"scheduledRedeploy,omitempty"`
	ScheduledForCertRedeploy corev1.ConditionStatus    `json:"scheduledCertRedeploy,omitempty"`
	UnderUpgrade             corev1.ConditionStatus    `json:"underUpgrade,omitempty"`
	UpgradePhase             ElasticsearchUpgradePhase `json:"upgradePhase,omitempty"`
}

type ClusterCondition struct {
	Type   ClusterConditionType   `json:"type"`
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type ClusterConditions []ClusterCondition

// The policy towards data redundancy to specify the number of redundant primary shards
//
// +kubebuilder:validation:Enum=FullRedundancy;MultipleRedundancy;SingleRedundancy;ZeroRedundancy
type RedundancyPolicyType string

const (
	// FullRedundancy - each index is fully replicated on every Data node in the cluster
	FullRedundancy RedundancyPolicyType = "FullRedundancy"
	// MultipleRedundancy - each index is spread over half of the Data nodes
	MultipleRedundancy RedundancyPolicyType = "MultipleRedundancy"
	// SingleRedundancy - one replica shard
	SingleRedundancy RedundancyPolicyType = "SingleRedundancy"
	// ZeroRedundancy - no replica shards
	ZeroRedundancy RedundancyPolicyType = "ZeroRedundancy"
)

// +kubebuilder:validation:Enum:=master;client;data
type ElasticsearchNodeRole string

const (
	ElasticsearchRoleClient ElasticsearchNodeRole = "client"
	ElasticsearchRoleData   ElasticsearchNodeRole = "data"
	ElasticsearchRoleMaster ElasticsearchNodeRole = "master"
)

type ShardAllocationState string

const (
	ShardAllocationAll       ShardAllocationState = "all"
	ShardAllocationNone      ShardAllocationState = "none"
	ShardAllocationPrimaries ShardAllocationState = "primaries"
	ShardAllocationUnknown   ShardAllocationState = "shard allocation unknown"
)

type PodStateMap map[PodStateType][]string

type PodStateType string

const (
	PodStateTypeReady    PodStateType = "ready"
	PodStateTypeNotReady PodStateType = "notReady"
	PodStateTypeFailed   PodStateType = "failed"
)

type ElasticsearchUpgradePhase string

const (
	NodeRestarting      ElasticsearchUpgradePhase = "nodeRestarting"
	RecoveringData      ElasticsearchUpgradePhase = "recoveringData"
	ControllerUpdated   ElasticsearchUpgradePhase = "controllerUpdated"
	PreparationComplete ElasticsearchUpgradePhase = "preparationComplete"
)

// Managed means that the operator is actively managing its resources and trying to keep the component active.
// It will only upgrade the component if it is safe to do so
// Unmanaged means that the operator will not take any action related to the component
//
// +kubebuilder:validation:Enum:=Managed;Unmanaged
type ManagementState string

const (
	ManagementStateManaged   ManagementState = "Managed"
	ManagementStateUnmanaged ManagementState = "Unmanaged"
)

// ClusterConditionType is a valid value for ClusterCondition.Type
type ClusterConditionType string

const (
	UpdatingSettings         ClusterConditionType = "UpdatingSettings"
	ScalingUp                ClusterConditionType = "ScalingUp"
	ScalingDown              ClusterConditionType = "ScalingDown"
	Restarting               ClusterConditionType = "Restarting"
	Recovering               ClusterConditionType = "Recovering"
	UpdatingESSettings       ClusterConditionType = "UpdatingESSettings"
	InvalidMasters           ClusterConditionType = "InvalidMasters"
	InvalidData              ClusterConditionType = "InvalidData"
	InvalidRedundancy        ClusterConditionType = "InvalidRedundancy"
	InvalidUUID              ClusterConditionType = "InvalidUUID"
	ESContainerWaiting       ClusterConditionType = "ElasticsearchContainerWaiting"
	ESContainerTerminated    ClusterConditionType = "ElasticsearchContainerTerminated"
	ProxyContainerWaiting    ClusterConditionType = "ProxyContainerWaiting"
	ProxyContainerTerminated ClusterConditionType = "ProxyContainerTerminated"
	Unschedulable            ClusterConditionType = "Unschedulable"
	NodeStorage              ClusterConditionType = "NodeStorage"
	CustomImage              ClusterConditionType = "CustomImageIgnored"
	DegradedState            ClusterConditionType = "Degraded"
	StorageClassName         ClusterConditionType = "StorageClassNameChangeIgnored"
	StorageSize              ClusterConditionType = "StorageSizeChangeIgnored"
	StorageStructure         ClusterConditionType = "StorageStructureChangeIgnored"
)
