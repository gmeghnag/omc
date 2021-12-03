package v1

import (
	"fmt"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A ServiceMeshMember object marks the namespace in which it lives as a member
// of the Service Mesh Control Plane referenced in the object.
// The ServiceMeshMember object should be created in each application namespace
// that must be part of the service mesh and must be named "default".
//
// When the ServiceMeshMember object is created, it causes the namespace to be
// added to the ServiceMeshMemberRoll within the namespace of the
// ServiceMeshControlPlane object the ServiceMeshMember references.
//
// To reference a ServiceMeshControlPlane, the user creating the ServiceMeshMember
// object must have the "use" permission on the referenced ServiceMeshControlPlane
// object. This permission is given via the mesh-users RoleBinding (and mesh-user
// Role) in the namespace of the referenced ServiceMeshControlPlane object.
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=smm,categories=maistra-io
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Control Plane",type="string",JSONPath=".status.annotations.controlPlaneRef",description="The ServiceMeshControlPlane this namespace belongs to"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description="Whether or not namespace is configured as a member of the mesh."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the object"
type ServiceMeshMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The desired state of this ServiceMeshMember.
	// +kubebuilder:validation:Required
	Spec ServiceMeshMemberSpec `json:"spec"`

	// The current status of this ServiceMeshMember. This data may be out of
	// date by some window of time.
	// +optional
	Status ServiceMeshMemberStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshMemberList contains a list of ServiceMeshMember objects
type ServiceMeshMemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshMember `json:"items"`
}

// ServiceMeshMemberSpec defines the member of the mesh
type ServiceMeshMemberSpec struct {

	// A reference to the ServiceMeshControlPlane object.
	ControlPlaneRef ServiceMeshControlPlaneRef `json:"controlPlaneRef"`
}

// ServiceMeshControlPlaneRef is a reference to a ServiceMeshControlPlane object
type ServiceMeshControlPlaneRef struct {

	// The name of the referenced ServiceMeshControlPlane object.
	Name string `json:"name"`

	// The namespace of the referenced ServiceMeshControlPlane object.
	Namespace string `json:"namespace"`
}

func (s ServiceMeshControlPlaneRef) String() string {
	return fmt.Sprintf("%s%c%s", s.Namespace, '/', s.Name)
}

// ServiceMeshMemberStatus represents the current state of a ServiceMeshMember.
type ServiceMeshMemberStatus struct {
	StatusBase `json:",inline"`

	// The generation observed by the controller during the most recent
	// reconciliation. The information in the status pertains to this particular
	// generation of the object.
	ObservedGeneration int64 `json:"observedGeneration"`

	// The generation of the ServiceMeshControlPlane object observed by the
	// controller during the most recent reconciliation of this
	// ServiceMeshMember.
	ServiceMeshGeneration int64 `json:"meshGeneration,omitempty"` // TODO: do we need this field at all?

	// The reconciled version of the ServiceMeshControlPlane object observed by
	// the controller during the most recent reconciliation of this
	// ServiceMeshMember.
	ServiceMeshReconciledVersion string `json:"meshReconciledVersion,omitempty"` // TODO: do we need this field at all?

	// Represents the latest available observations of a ServiceMeshMember's
	// current state.
	Conditions []ServiceMeshMemberCondition `json:"conditions"`
}

// ServiceMeshMemberConditionType represents the type of the condition.  Condition types are:
// Reconciled, NamespaceConfigured
type ServiceMeshMemberConditionType string

const (
	// ConditionTypeReconciled signifies whether or not the controller has
	// updated the ServiceMeshMemberRoll object based on this ServiceMeshMember.
	ConditionTypeMemberReconciled ServiceMeshMemberConditionType = "Reconciled"
	// ConditionTypeReady signifies whether the namespace has been configured
	// to use the mesh
	ConditionTypeMemberReady ServiceMeshMemberConditionType = "Ready" // TODO: remove the Ready condition in v2
)

type ServiceMeshMemberConditionReason string

const (
	// ConditionReasonDeletionError ...
	ConditionReasonMemberCannotCreateMemberRoll          ServiceMeshMemberConditionReason = "CreateMemberRollFailed"
	ConditionReasonMemberCannotUpdateMemberRoll          ServiceMeshMemberConditionReason = "UpdateMemberRollFailed"
	ConditionReasonMemberCannotDeleteMemberRoll          ServiceMeshMemberConditionReason = "DeleteMemberRollFailed"
	ConditionReasonMemberNamespaceNotExists              ServiceMeshMemberConditionReason = "NamespaceNotExists"
	ConditionReasonMemberReferencesDifferentControlPlane ServiceMeshMemberConditionReason = "ReferencesDifferentControlPlane"
	ConditionReasonMemberTerminating                     ServiceMeshMemberConditionReason = "Terminating"
)

// Condition represents a specific condition on a resource
type ServiceMeshMemberCondition struct {
	Type               ServiceMeshMemberConditionType   `json:"type,omitempty"`
	Status             core.ConditionStatus             `json:"status,omitempty"`
	LastTransitionTime metav1.Time                      `json:"lastTransitionTime,omitempty"`
	Reason             ServiceMeshMemberConditionReason `json:"reason,omitempty"`
	Message            string                           `json:"message,omitempty"`
}

// GetCondition removes a condition for the list of conditions
func (s *ServiceMeshMemberStatus) GetCondition(conditionType ServiceMeshMemberConditionType) ServiceMeshMemberCondition {
	if s == nil {
		return ServiceMeshMemberCondition{Type: conditionType, Status: core.ConditionUnknown}
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			return s.Conditions[i]
		}
	}
	return ServiceMeshMemberCondition{Type: conditionType, Status: core.ConditionUnknown}
}

// SetCondition sets a specific condition in the list of conditions
func (s *ServiceMeshMemberStatus) SetCondition(condition ServiceMeshMemberCondition) *ServiceMeshMemberStatus {
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
