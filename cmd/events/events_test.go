package events

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterOnEventType(t *testing.T) {
	tests := []struct {
		name     string
		selector []string
		expected []string
	}{
		{
			name:     "Basic match on the event type",
			selector: []string{"Normal"},
			expected: []string{"test2"},
		},
		{
			name:     "Selector and event type have different case",
			selector: []string{"warning"},
			expected: []string{"test1"},
		},
		{
			name:     "Selector shouldn't match anything",
			selector: []string{"NoMatch"},
			expected: []string{},
		},
		{
			name:     "Match using multiple selectors",
			selector: []string{"Warning", "Normal"},
			expected: []string{"test1", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testData := corev1.EventList{
				Items: []corev1.Event{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test1",
							Namespace: "testns",
						},
						Type: "Warning",
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test2",
							Namespace: "testns",
						},
						Type: "Normal",
					},
				}}
			FilterEventList(&testData, tt.selector, "")
			actual := []string{}
			for _, event := range testData.Items {
				actual = append(actual, event.Name)
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestFilterOnResource(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		expected []string
	}{
		{
			name:     "Basic match on pod resource",
			selector: "pod/testpod",
			expected: []string{"test1"},
		},
		{
			name:     "Match on capitalized resource kind",
			selector: "Pod/testpod",
			expected: []string{"test1"},
		},
		{
			name:     "Match on capitalized resource name",
			selector: "pod/Testpod",
			expected: []string{"test1"},
		},
		{
			name:     "Combination of resource kind and name has no matches",
			selector: "pod/testmcp",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testData := corev1.EventList{
				Items: []corev1.Event{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test1",
							Namespace: "testns",
						},
						InvolvedObject: corev1.ObjectReference{
							Kind:       "Pod",
							Name:       "testpod",
							APIVersion: "v1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test2",
							Namespace: "testns",
						},
						InvolvedObject: corev1.ObjectReference{
							Kind:       "MachineConfigPool",
							Name:       "testmcp",
							APIVersion: "machineconfiguration.openshift.io/v1",
						},
					},
				}}
			FilterEventList(&testData, []string{}, tt.selector)
			actual := []string{}
			for _, event := range testData.Items {
				actual = append(actual, event.Name)
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}
