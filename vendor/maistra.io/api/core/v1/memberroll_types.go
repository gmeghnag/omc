package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The ServiceMeshMemberRoll object configures which namespaces belong to a
// service mesh. Only namespaces listed in the ServiceMeshMemberRoll will be
// affected by the control plane. Any number of namespaces can be added, but a
// namespace may not exist in more than one service mesh. The
// ServiceMeshMemberRoll object must be created in the same namespace as
// the ServiceMeshControlPlane object and must be named "default".
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=smmr,categories=maistra-io
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.annotations.configuredMemberCount",description="How many of the total number of member namespaces are configured"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason",description="Whether all member namespaces have been configured or why that's not the case"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the object"
// +kubebuilder:printcolumn:name="Members",type="string",JSONPath=".status.members",description="Namespaces that are members of this Control Plane",priority=1
type ServiceMeshMemberRoll struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired list of members of the service mesh.
	// +kubebuilder:validation:Required
	Spec ServiceMeshMemberRollSpec `json:"spec"`

	// The current status of this ServiceMeshMemberRoll. This data may be out
	// of date by some window of time.
	Status ServiceMeshMemberRollStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshMemberRollList contains a list of ServiceMeshMemberRoll
type ServiceMeshMemberRollList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshMemberRoll `json:"items"`
}

// ServiceMeshMemberRollSpec is the specification of the desired list of
// members of the service mesh.
type ServiceMeshMemberRollSpec struct {

	//  List of namespaces that should be members of the service mesh.
	// +optional
	// +nullable
	Members []string `json:"members,omitempty"`
}

// ServiceMeshMemberRollStatus represents the current state of a ServiceMeshMemberRoll.
type ServiceMeshMemberRollStatus struct {
	StatusBase `json:",inline"`

	// The generation observed by the controller during the most recent
	// reconciliation. The information in the status pertains to this particular
	// generation of the object.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// The generation of the ServiceMeshControlPlane object observed by the
	// controller during the most recent reconciliation of this
	// ServiceMeshMemberRoll.
	ServiceMeshGeneration int64 `json:"meshGeneration,omitempty"`

	// The reconciled version of the ServiceMeshControlPlane object observed by
	// the controller during the most recent reconciliation of this
	// ServiceMeshMemberRoll.
	ServiceMeshReconciledVersion string `json:"meshReconciledVersion,omitempty"`

	// Complete list of namespaces that are configured as members of the service
	// mesh	- this includes namespaces specified in spec.members and those that
	// contain a ServiceMeshMember object
	// +optional
	// +nullable
	Members []string `json:"members"`

	// List of namespaces that are configured as members of the service mesh.
	// +optional
	// +nullable
	ConfiguredMembers []string `json:"configuredMembers"`

	// List of namespaces that haven't been configured as members of the service
	// mesh yet.
	// +optional
	// +nullable
	PendingMembers []string `json:"pendingMembers"`

	// List of namespaces that are being removed as members of the service
	// mesh.
	// +optional
	// +nullable
	TerminatingMembers []string `json:"terminatingMembers"`

	// Represents the latest available observations of this ServiceMeshMemberRoll's
	// current state.
	// +optional
	// +nullable
	Conditions []ServiceMeshMemberRollCondition `json:"conditions"`

	// Represents the latest available observations of each member's
	// current state.
	// +optional
	// +nullable
	MemberStatuses []ServiceMeshMemberStatusSummary `json:"memberStatuses"`
}

// ServiceMeshMemberStatusSummary represents a summary status of a ServiceMeshMember.
type ServiceMeshMemberStatusSummary struct {
	Namespace  string                       `json:"namespace"`
	Conditions []ServiceMeshMemberCondition `json:"conditions"`
}

// ServiceMeshMemberRollConditionType represents the type of the condition.  Condition types are:
// Reconciled, NamespaceConfigured
type ServiceMeshMemberRollConditionType string

const (
	// ConditionTypeMemberRollReady signifies whether the namespace has been configured
	// to use the mesh
	ConditionTypeMemberRollReady ServiceMeshMemberRollConditionType = "Ready"
)

type ServiceMeshMemberRollConditionReason string

const (
	// ConditionReasonConfigured indicates that all namespaces were configured
	ServiceMeshMemberRollConditionReasonConditionReasonConfigured ServiceMeshMemberRollConditionReason = "Configured"
	// ConditionReasonReconcileError indicates that one of the namespaces to configure could not be configured
	ServiceMeshMemberRollConditionReasonConditionReasonReconcileError ServiceMeshMemberRollConditionReason = "ReconcileError"
	// ConditionReasonSMCPMissing indicates that the ServiceMeshControlPlane resource does not exist
	ServiceMeshMemberRollConditionReasonConditionReasonSMCPMissing ServiceMeshMemberRollConditionReason = "ErrSMCPMissing"
	// ConditionReasonMultipleSMCP indicates that multiple ServiceMeshControlPlane resources exist in the namespace
	ServiceMeshMemberRollConditionReasonConditionReasonMultipleSMCP ServiceMeshMemberRollConditionReason = "ErrMultipleSMCPs"
	// ConditionReasonSMCPNotReconciled indicates that reconciliation of the SMMR was skipped because the SMCP has not been reconciled
	ServiceMeshMemberRollConditionReasonConditionReasonSMCPNotReconciled ServiceMeshMemberRollConditionReason = "SMCPReconciling"
)

// Condition represents a specific condition on a resource
type ServiceMeshMemberRollCondition struct {
	Type               ServiceMeshMemberRollConditionType   `json:"type,omitempty"`
	Status             core.ConditionStatus                 `json:"status,omitempty"`
	LastTransitionTime metav1.Time                          `json:"lastTransitionTime,omitempty"`
	Reason             ServiceMeshMemberRollConditionReason `json:"reason,omitempty"`
	Message            string                               `json:"message,omitempty"`
}

// GetCondition removes a condition for the list of conditions
func (s *ServiceMeshMemberRollStatus) GetCondition(conditionType ServiceMeshMemberRollConditionType) ServiceMeshMemberRollCondition {
	if s == nil {
		return ServiceMeshMemberRollCondition{Type: conditionType, Status: core.ConditionUnknown}
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			return s.Conditions[i]
		}
	}
	return ServiceMeshMemberRollCondition{Type: conditionType, Status: core.ConditionUnknown}
}

// SetCondition sets a specific condition in the list of conditions
func (s *ServiceMeshMemberRollStatus) SetCondition(condition ServiceMeshMemberRollCondition) *ServiceMeshMemberRollStatus {
	if s == nil {
		return nil
	}
	now := metav1.Now()
	for i := range s.Conditions {
		if s.Conditions[i].Type == condition.Type {
			if s.Conditions[i].Status != condition.Status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = s.Conditions[i].LastTransitionTime
			}
			s.Conditions[i] = condition
			return s
		}
	}

	// If the condition does not exist,
	// initialize the lastTransitionTime
	condition.LastTransitionTime = now
	s.Conditions = append(s.Conditions, condition)
	return s
}
