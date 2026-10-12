package podstatus

import (
	"strings"

	core "k8s.io/kubernetes/pkg/apis/core"
)

// CorrectDisplayStatus fixes misleading Init:* status values produced by the
// kubectl pod printer when initContainerStatuses in a must-gather snapshot are
// incomplete but the pod is fully initialized and running.
func CorrectDisplayStatus(pod *core.Pod, displayedReason string) string {
	if pod == nil || !strings.HasPrefix(displayedReason, "Init:") {
		return displayedReason
	}
	if pod.Status.Phase != core.PodRunning {
		return displayedReason
	}
	if !isPodInitialized(pod.Status.Conditions) {
		return displayedReason
	}
	if hasPodReadyCondition(pod.Status.Conditions) {
		return "Running"
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Running != nil {
			return "NotReady"
		}
	}
	return displayedReason
}

func isPodInitialized(conditions []core.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == core.PodInitialized {
			return condition.Status == core.ConditionTrue
		}
	}
	return false
}

func hasPodReadyCondition(conditions []core.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == core.PodReady {
			return condition.Status == core.ConditionTrue
		}
	}
	return false
}
