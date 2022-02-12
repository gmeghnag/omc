package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IndexManagementSpec specifies index management for an Elasticsearch cluster
// +k8s:openapi-gen=true
type IndexManagementSpec struct {
	// A list of polices for managing an indices
	//
	// +optional
	Policies []IndexManagementPolicySpec `json:"policies"`

	// Mappings of policies to indicies
	//
	// +optional
	Mappings []IndexManagementPolicyMappingSpec `json:"mappings"`
}

// TimeUnit is a time unit like h,m,d
//
// +kubebuilder:validation:Pattern:="^([0-9]+)([yMwdhHms]{0,1})$"
type TimeUnit string

// IndexManagementPolicySpec is a definition of an index management policy
// +k8s:openapi-gen=true
type IndexManagementPolicySpec struct {
	// The unique name of the policy
	Name string `json:"name"`

	// How often to check an index meets the desired criteria (e.g. 1m)
	PollInterval TimeUnit `json:"pollInterval"`

	Phases IndexManagementPhasesSpec `json:"phases"`
}

// +k8s:openapi-gen=true
type IndexManagementPhasesSpec struct {
	// +nullable
	Hot *IndexManagementHotPhaseSpec `json:"hot,omitempty"`
	// +nullable
	Delete *IndexManagementDeletePhaseSpec `json:"delete,omitempty"`
}

// +k8s:openapi-gen=true
type IndexManagementDeletePhaseSpec struct {
	// The minimum age of an index before it should be deleted (e.g. 10d)
	MinAge TimeUnit `json:"minAge"`

	// How often to run a new prune-namespaces job
	// +optional
	PruneNamespacesInterval TimeUnit `json:"pruneNamespacesInterval,omitempty"`

	// The per namespace specification to delete documents older than a given minimum age
	// +optional
	Namespaces []IndexManagementDeleteNamespaceSpec `json:"namespaceSpec,omitempty"`
}

type IndexManagementDeleteNamespaceSpec struct {
	// Target Namespace to delete logs older than MinAge (defaults to 7d)
	// Can be one namespace name or a prefix (e.g., "openshift-" covers all namespaces with this prefix)
	Namespace string `json:"namespace"`

	// Delete the records matching the namespaces which are older than this MinAge (e.g. 1d)
	// +optional
	MinAge TimeUnit `json:"minAge,omitempty"`
}

// +k8s:openapi-gen=true
type IndexManagementHotPhaseSpec struct {
	// +optional
	Actions IndexManagementActionsSpec `json:"actions"`
}

// +k8s:openapi-gen=true
type IndexManagementActionsSpec struct {
	// +nullable
	// +optional
	Rollover *IndexManagementActionSpec `json:"rollover"`
}

// +k8s:openapi-gen=true
type IndexManagementActionSpec struct {
	// The maximum age of an index before it should be rolled over (e.g. 7d)
	MaxAge TimeUnit `json:"maxAge"`
}

// IndexManagementPolicyMappingSpec maps a management policy to an index
// +k8s:openapi-gen=true
type IndexManagementPolicyMappingSpec struct {
	// The unique name of the policy mapping
	//
	// +optional
	Name string `json:"name"`

	// A reference to a defined policy
	// +optional
	PolicyRef string `json:"policyRef"`

	// Aliases to apply to a template
	Aliases []string `json:"aliases,omitempty"`
}

type PolicyMap map[string]IndexManagementPolicySpec

func (spec *IndexManagementSpec) PolicyMap() PolicyMap {
	policyMap := map[string]IndexManagementPolicySpec{}
	for _, spec := range spec.Policies {
		policyMap[spec.Name] = spec
	}
	return policyMap
}

func (policyMap *PolicyMap) HasPolicy(name string) bool {
	_, found := map[string]IndexManagementPolicySpec(*policyMap)[name]
	return found
}

// +k8s:openapi-gen=true
type IndexManagementStatus struct {
	State       IndexManagementState           `json:"state,omitempty"`
	Reason      IndexManagementStatusReason    `json:"reason,omitempty"`
	Message     string                         `json:"message,omitempty"`
	LastUpdated metav1.Time                    `json:"lastUpdated,omitempty"`
	Policies    []IndexManagementPolicyStatus  `json:"policies,omitempty"`
	Mappings    []IndexManagementMappingStatus `json:"mappings,omitempty"`
}

func NewIndexManagementStatus() *IndexManagementStatus {
	return &IndexManagementStatus{
		State:       IndexManagementStateAccepted,
		Reason:      IndexManagementStatusReasonPassed,
		LastUpdated: metav1.Now(),
	}
}

// IndexManagementState of IndexManagment
type IndexManagementState string

const (
	// IndexManagementStateAccepted when polices and mappings are well defined and pass validations
	IndexManagementStateAccepted IndexManagementState = "Accepted"

	// IndexManagementStateDegraded some polices and mappings have failed validations
	IndexManagementStateDegraded IndexManagementState = "Degraded"

	// IndexManagementStateDropped when IndexManagement is not defined or there are no valid polices and mappings
	IndexManagementStateDropped IndexManagementState = "Dropped"
)

type IndexManagementStatusReason string

const (
	IndexManagementStatusReasonPassed           = "PassedValidation"
	IndexManagementStatusReasonUndefined        = "Undefined"
	IndexManagementStatusReasonValidationFailed = "OneOrMoreValidationsFailed"
)

type IndexManagementMappingStatus struct {
	// Name of the corresponding mapping for this status
	Name string `json:"name,omitempty"`

	// State of the corresponding mapping for this status
	State IndexManagementMappingState `json:"state,omitempty"`

	Reason IndexManagementMappingReason `json:"reason,omitempty"`

	Message string `json:"message,omitempty"`

	// Reasons for the state of the corresponding mapping for this status
	Conditions []IndexManagementMappingCondition `json:"conditions,omitempty"`

	// LastUpdated represents the last time that the status was updated.
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

func NewIndexManagementMappingStatus(name string) *IndexManagementMappingStatus {
	return &IndexManagementMappingStatus{
		Name:        name,
		State:       IndexManagementMappingStateAccepted,
		Reason:      IndexManagementMappingReasonConditionsMet,
		LastUpdated: metav1.Now(),
	}
}

func (status *IndexManagementMappingStatus) AddPolicyMappingCondition(conditionType IndexManagementMappingConditionType, reason IndexManagementMappingConditionReason, message string) {
	status.Conditions = append(status.Conditions, IndexManagementMappingCondition{
		Type:    conditionType,
		Reason:  reason,
		Status:  corev1.ConditionFalse,
		Message: message,
	})
}

type IndexManagementMappingState string

const (
	// IndexManagementMappingStateAccepted passes validations
	IndexManagementMappingStateAccepted IndexManagementMappingState = "Accepted"

	// IndexManagementMappingStateDropped fails validations
	IndexManagementMappingStateDropped IndexManagementMappingState = "Dropped"
)

type IndexManagementMappingReason string

const (
	IndexManagementMappingReasonConditionsMet    IndexManagementMappingReason = "ConditionsMet"
	IndexManagementMappingReasonConditionsNotMet IndexManagementMappingReason = "ConditionsNotMet"
)

type IndexManagementMappingCondition struct {
	Type    IndexManagementMappingConditionType   `json:"type,omitempty"`
	Reason  IndexManagementMappingConditionReason `json:"reason,omitempty"`
	Status  corev1.ConditionStatus                `json:"status,omitempty"`
	Message string                                `json:"message,omitempty"`
}

type IndexManagementMappingConditionType string

const (
	IndexManagementMappingConditionTypeName      IndexManagementMappingConditionType = "Name"
	IndexManagementMappingConditionTypePolicyRef IndexManagementMappingConditionType = "PolicyRef"
)

type IndexManagementMappingConditionReason string

const (
	IndexManagementMappingReasonMissing   IndexManagementMappingConditionReason = "Missing"
	IndexManagementMappingReasonNonUnique IndexManagementMappingConditionReason = "NonUnique"
)

type IndexManagementPolicyStatus struct {
	// Name of the corresponding policy for this status
	Name string `json:"name,omitempty"`

	// State of the corresponding policy for this status
	State IndexManagementPolicyState `json:"state,omitempty"`

	// Reasons for the state of the corresponding policy for this status
	Reason IndexManagementPolicyReason `json:"reason,omitempty"`

	// Message about the corresponding policy
	Message string `json:"message,omitempty"`

	// Reasons for the state of the corresponding policy for this status
	Conditions []IndexManagementPolicyCondition `json:"conditions,omitempty"`

	// LastUpdated represents the last time that the status was updated.
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

func NewIndexManagementPolicyStatus(name string) *IndexManagementPolicyStatus {
	return &IndexManagementPolicyStatus{
		Name:        name,
		State:       IndexManagementPolicyStateAccepted,
		Reason:      IndexManagementPolicyReasonConditionsMet,
		LastUpdated: metav1.Now(),
	}
}

func (status *IndexManagementPolicyStatus) AddPolicyCondition(conditionType IndexManagementPolicyConditionType, reason IndexManagementPolicyConditionReason, message string) {
	status.Conditions = append(status.Conditions, IndexManagementPolicyCondition{
		Type:    conditionType,
		Reason:  reason,
		Status:  corev1.ConditionFalse,
		Message: message,
	})
}

type IndexManagementPolicyState string

const (
	// IndexManagementPolicyStateAccepted passes validations
	IndexManagementPolicyStateAccepted IndexManagementPolicyState = "Accepted"

	// IndexManagementPolicyStateDropped fails validations
	IndexManagementPolicyStateDropped IndexManagementPolicyState = "Dropped"
)

type IndexManagementPolicyReason string

const (
	IndexManagementPolicyReasonConditionsMet    IndexManagementPolicyReason = "ConditionsMet"
	IndexManagementPolicyReasonConditionsNotMet IndexManagementPolicyReason = "ConditionsNotMet"
)

type IndexManagementPolicyCondition struct {
	Type    IndexManagementPolicyConditionType   `json:"type,omitempty"`
	Reason  IndexManagementPolicyConditionReason `json:"reason,omitempty"`
	Status  corev1.ConditionStatus               `json:"status,omitempty"`
	Message string                               `json:"message,omitempty"`
}

type IndexManagementPolicyConditionType string

const (
	IndexManagementPolicyConditionTypeName         IndexManagementPolicyConditionType = "Name"
	IndexManagementPolicyConditionTypePollInterval IndexManagementPolicyConditionType = "PollInterval"
	IndexManagementPolicyConditionTypeTimeUnit     IndexManagementPolicyConditionType = "TimeUnit"
)

type IndexManagementPolicyConditionReason string

const (
	IndexManagementPolicyReasonMalformed IndexManagementPolicyConditionReason = "MalFormed"
	IndexManagementPolicyReasonMissing   IndexManagementPolicyConditionReason = "Missing"
	IndexManagementPolicyReasonNonUnique IndexManagementPolicyConditionReason = "NonUnique"
)
