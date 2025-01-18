package events

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
)

func GetLastTime(event corev1.Event) metav1.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp
	}
	return event.GetCreationTimestamp()
}

func convertType(in *corev1.EventList, out *api.EventList) {
	out.Items = make([]api.Event, len(in.Items))
	for i := range in.Items {
		out.Items[i].Type = in.Items[i].Type
		out.Items[i].Reason = in.Items[i].Reason
		out.Items[i].Message = in.Items[i].Message
		out.Items[i].InvolvedObject.Kind = in.Items[i].InvolvedObject.Kind
		out.Items[i].InvolvedObject.Name = in.Items[i].InvolvedObject.Name
	}
}
