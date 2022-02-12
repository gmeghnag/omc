package v1

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/status"

	corev1 "k8s.io/api/core/v1"
)

// Aliases for convenience
type Condition = status.Condition
type ConditionType = status.ConditionType
type ConditionReason = status.ConditionReason
type Conditions = status.Conditions

func NewConditions(c ...Condition) Conditions { return status.NewConditions(c...) }

func NewCondition(t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) Condition {
	return Condition{Type: t, Status: s, Reason: r, Message: fmt.Sprintf(format, args...)}
}

const (
	// Ready indicates the service is ready.
	//
	// Ready=True means the operands are running and providing some service.
	// See the Degraded condition to distinguish full service from partial service.
	//
	// Ready=False means the operands cannot provide any service, and
	// the operator cannot recover without some external change. Either
	// the spec is invalid, or there is some environmental problem that is
	// outside of the the operator's control.
	//
	// Ready=Unknown means the operator is in transition.
	//
	ConditionReady status.ConditionType = "Ready"

	// Degraded indicates partial service is available.
	//
	// Degraded=True means the operands can fulfill some of the `spec`, but not all,
	// even when Ready=True.
	//
	// Degraded=False with Ready=True means the operands are providing full service.
	//
	// Degraded=Unknown means the operator is in transition.
	//
	ConditionDegraded status.ConditionType = "Degraded"
)

const (
	// Invalid spec is ill-formed in some way, or contains unknown references.
	ReasonInvalid status.ConditionReason = "Invalid"
	// MissingResources spec refers to resources that can't be located.
	ReasonMissingResource status.ConditionReason = "MissingResource"
	// Unused spec defines a valid object but it is never used.
	ReasonUnused status.ConditionReason = "Unused"
	// Connecting object is unready because a connection is in progress.
	ReasonConnecting status.ConditionReason = "Connecting"
)

// SetCondition returns true if the condition changed or is new.
func SetCondition(cs *status.Conditions, t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) bool {
	return cs.SetCondition(NewCondition(t, s, r, format, args...))
}

type NamedConditions map[string]status.Conditions

func (nc NamedConditions) Set(name string, cond status.Condition) bool {
	conds := nc[name]
	ret := conds.SetCondition(cond)
	nc[name] = conds
	return ret
}

func (nc NamedConditions) SetCondition(name string, t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, format string, args ...interface{}) bool {
	return nc.Set(name, NewCondition(t, s, r, format, args...))
}

func (nc NamedConditions) IsAllReady() bool {
	for _, conditions := range nc {
		for _, cond := range conditions {
			if cond.Type == ConditionReady && cond.IsFalse() {
				return false
			}
		}
	}
	return true
}
