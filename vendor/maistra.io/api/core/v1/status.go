package v1

import (
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type StatusBase struct {
	// Annotations is an unstructured key value map used to store additional,
	// usually redundant status information, such as the number of components
	// deployed by the ServiceMeshControlPlane (number is redundant because
	// you could just as easily count the elements in the ComponentStatus
	// array). The reason to add this redundant information is to make it
	// available to kubectl, which does not yet allow counting objects in
	// JSONPath expressions.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (s *StatusBase) GetAnnotation(name string) string {
	if s.Annotations == nil {
		return ""
	}
	return s.Annotations[name]
}

func (s *StatusBase) SetAnnotation(name string, value string) {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}
	s.Annotations[name] = value
}

func (s *StatusBase) RemoveAnnotation(name string) {
	if s.Annotations != nil {
		delete(s.Annotations, name)
	}
}

// StatusType represents the status for a control plane, component, or resource
type StatusType struct {

	// Represents the latest available observations of the object's current state.
	Conditions []Condition `json:"conditions,omitempty"`
}

// NewStatus returns a new StatusType object
func NewStatus() StatusType {
	return StatusType{Conditions: make([]Condition, 0, 3)}
}

type ComponentStatusList struct {
	//+optional
	ComponentStatus []ComponentStatus `json:"components,omitempty"`
}

// FindComponentByName returns the status for a specific component
func (s *ComponentStatusList) FindComponentByName(name string) *ComponentStatus {
	for i, status := range s.ComponentStatus {
		if status.Resource == name {
			return &s.ComponentStatus[i]
		}
	}
	return nil
}

// NewComponentStatus returns a new ComponentStatus object
func NewComponentStatus() *ComponentStatus {
	return &ComponentStatus{StatusType: NewStatus()}
}

// ComponentStatus represents the status of an object with children
type ComponentStatus struct {
	StatusType `json:",inline"`

	// The name of the component this status pertains to.
	Resource string `json:"resource,omitempty"`

	// TODO: can we remove this? it's not used anywhere
	// The status of each resource that comprises this component.
	Resources []*StatusType `json:"children,omitempty"`
}

// ConditionType represents the type of the condition.  Condition stages are:
// Installed, Reconciled, Ready
type ConditionType string

const (
	// ConditionTypeInstalled signifies the whether or not the controller has
	// installed the resources defined through the CR.
	ConditionTypeInstalled ConditionType = "Installed"
	// ConditionTypeReconciled signifies the whether or not the controller has
	// reconciled the resources defined through the CR.
	ConditionTypeReconciled ConditionType = "Reconciled"
	// ConditionTypeReady signifies the whether or not any Deployment, StatefulSet,
	// etc. resources are Ready.
	ConditionTypeReady ConditionType = "Ready"
)

// ConditionStatus represents the status of the condition
type ConditionStatus string

const (
	// ConditionStatusTrue represents completion of the condition, e.g.
	// Initialized=True signifies that initialization has occurred.
	ConditionStatusTrue ConditionStatus = "True"
	// ConditionStatusFalse represents incomplete status of the condition, e.g.
	// Initialized=False signifies that initialization has not occurred or has
	// failed.
	ConditionStatusFalse ConditionStatus = "False"
	// ConditionStatusUnknown represents unknown completion of the condition, e.g.
	// Initialized=Unknown signifies that initialization may or may not have been
	// completed.
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

// ConditionReason represents a short message indicating how the condition came
// to be in its present state.
type ConditionReason string

const (
	// ConditionReasonDeletionError ...
	ConditionReasonDeletionError ConditionReason = "DeletionError"
	// ConditionReasonInstallSuccessful ...
	ConditionReasonInstallSuccessful ConditionReason = "InstallSuccessful"
	// ConditionReasonInstallError ...
	ConditionReasonInstallError ConditionReason = "InstallError"
	// ConditionReasonReconcileSuccessful ...
	ConditionReasonReconcileSuccessful ConditionReason = "ReconcileSuccessful"
	// ConditionReasonValidationError ...
	ConditionReasonValidationError ConditionReason = "ValidationError"
	// ConditionReasonDependencyMissingError ...
	ConditionReasonDependencyMissingError ConditionReason = "DependencyMissingError"
	// ConditionReasonReconcileError ...
	ConditionReasonReconcileError ConditionReason = "ReconcileError"
	// ConditionReasonResourceCreated ...
	ConditionReasonResourceCreated ConditionReason = "ResourceCreated"
	// ConditionReasonSpecUpdated ...
	ConditionReasonSpecUpdated ConditionReason = "SpecUpdated"
	// ConditionReasonUpdateSuccessful ...
	ConditionReasonUpdateSuccessful ConditionReason = "UpdateSuccessful"
	// ConditionReasonComponentsReady ...
	ConditionReasonComponentsReady ConditionReason = "ComponentsReady"
	// ConditionReasonComponentsNotReady ...
	ConditionReasonComponentsNotReady ConditionReason = "ComponentsNotReady"
	// ConditionReasonProbeError ...
	ConditionReasonProbeError ConditionReason = "ProbeError"
	// ConditionReasonPausingInstall ...
	ConditionReasonPausingInstall ConditionReason = "PausingInstall"
	// ConditionReasonPausingUpdate ...
	ConditionReasonPausingUpdate ConditionReason = "PausingUpdate"
	// ConditionReasonDeleting ...
	ConditionReasonDeleting ConditionReason = "Deleting"
	// ConditionReasonDeleted ...
	ConditionReasonDeleted ConditionReason = "Deleted"
)

// A Condition represents a specific observation of the object's state.
type Condition struct {

	// The type of this condition.
	Type ConditionType `json:"type,omitempty"`

	// The status of this condition. Can be True, False or Unknown.
	Status ConditionStatus `json:"status,omitempty"`

	// Unique, single-word, CamelCase reason for the condition's last transition.
	Reason ConditionReason `json:"reason,omitempty"`

	// Human-readable message indicating details about the last transition.
	Message string `json:"message,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

func (c *Condition) Matches(status ConditionStatus, reason ConditionReason, message string) bool {
	return c.Status == status && c.Reason == reason && c.Message == message
}

// ComposeReconciledVersion returns a string for use in ReconciledVersion fields
func ComposeReconciledVersion(operatorVersion string, generation int64) string {
	return fmt.Sprintf("%s-%d", operatorVersion, generation)
}

// GetCondition removes a condition for the list of conditions
func (s *StatusType) GetCondition(conditionType ConditionType) Condition {
	if s == nil {
		return Condition{Type: conditionType, Status: ConditionStatusUnknown}
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			return s.Conditions[i]
		}
	}
	return Condition{Type: conditionType, Status: ConditionStatusUnknown}
}

// SetCondition sets a specific condition in the list of conditions
func (s *StatusType) SetCondition(condition Condition) *StatusType {
	if s == nil {
		return nil
	}
	// These only get serialized out to the second.  This can break update
	// skipping, as the time in the resource returned from the client may not
	// match the time in our cached status during a reconcile.  We truncate here
	// to save any problems down the line.
	now := metav1.NewTime(time.Now().Truncate(time.Second))
	for i, prevCondition := range s.Conditions {
		if prevCondition.Type == condition.Type {
			if prevCondition.Status != condition.Status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = prevCondition.LastTransitionTime
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

// RemoveCondition removes a condition for the list of conditions
func (s *StatusType) RemoveCondition(conditionType ConditionType) *StatusType {
	if s == nil {
		return nil
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == conditionType {
			s.Conditions = append(s.Conditions[:i], s.Conditions[i+1:]...)
			return s
		}
	}
	return s
}

// ResourceKey is a typedef for key used in ManagedGenerations.  It is a string
// with the format: namespace/name=group/version,kind
type ResourceKey string

// NewResourceKey for the object and type
func NewResourceKey(o metav1.Object, t metav1.Type) ResourceKey {
	return ResourceKey(fmt.Sprintf("%s/%s=%s,Kind=%s", o.GetNamespace(), o.GetName(), t.GetAPIVersion(), t.GetKind()))
}

// ToUnstructured returns a an Unstructured object initialized with Namespace,
// Name, APIVersion, and Kind fields from the ResourceKey
func (key ResourceKey) ToUnstructured() *unstructured.Unstructured {
	// ResourceKey is guaranteed to be at least "/=," meaning we are guaranteed
	// to get two elements in all of the splits
	retval := &unstructured.Unstructured{}
	parts := strings.SplitN(string(key), "=", 2)
	nn := strings.SplitN(parts[0], "/", 2)
	gvk := strings.SplitN(parts[1], ",Kind=", 2)
	retval.SetNamespace(nn[0])
	retval.SetName(nn[1])
	retval.SetAPIVersion(gvk[0])
	retval.SetKind(gvk[1])
	return retval
}
