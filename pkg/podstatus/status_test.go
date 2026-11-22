package podstatus

import (
	"testing"

	core "k8s.io/kubernetes/pkg/apis/core"
)

func TestCorrectDisplayStatus(t *testing.T) {
	runningPod := func(ready bool) *core.Pod {
		pod := &core.Pod{
			Status: core.PodStatus{
				Phase: core.PodRunning,
				Conditions: []core.PodCondition{
					{Type: core.PodInitialized, Status: core.ConditionTrue},
				},
				ContainerStatuses: []core.ContainerStatus{
					{Name: "app", State: core.ContainerState{Running: &core.ContainerStateRunning{}}},
				},
			},
		}
		if ready {
			pod.Status.Conditions = append(pod.Status.Conditions, core.PodCondition{
				Type: core.PodReady, Status: core.ConditionTrue,
			})
		}
		return pod
	}

	tests := []struct {
		name   string
		pod    *core.Pod
		reason string
		want   string
	}{
		{
			name:   "initialized running pod with misleading init status",
			pod:    runningPod(true),
			reason: "Init:0/3",
			want:   "Running",
		},
		{
			name:   "initialized but not ready",
			pod:    runningPod(false),
			reason: "Init:1/3",
			want:   "NotReady",
		},
		{
			name:   "still initializing",
			pod: &core.Pod{
				Status: core.PodStatus{
					Phase: core.PodPending,
					Conditions: []core.PodCondition{
						{Type: core.PodInitialized, Status: core.ConditionFalse},
					},
				},
			},
			reason: "Init:0/3",
			want:   "Init:0/3",
		},
		{
			name:   "non init display reason unchanged",
			pod:    runningPod(true),
			reason: "CrashLoopBackOff",
			want:   "CrashLoopBackOff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CorrectDisplayStatus(tt.pod, tt.reason); got != tt.want {
				t.Fatalf("CorrectDisplayStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
