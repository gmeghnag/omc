package events

import (
	"os"
	"path/filepath"
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

// writeEvents writes a core events.yaml for one namespace of a must-gather root.
func writeEvents(t *testing.T, root, namespace string, events []ev) {
	t.Helper()
	dir := filepath.Join(root, "namespaces", namespace, "core")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var items string
	for _, e := range events {
		items += "- apiVersion: v1\n  kind: Event\n  metadata:\n    name: " + e.name +
			"\n    namespace: " + namespace + "\n    uid: " + e.uid +
			"\n  reason: R\n  type: Normal\n"
	}
	body := "apiVersion: v1\nkind: EventList\nitems:\n" + items
	if err := os.WriteFile(filepath.Join(dir, "events.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

type ev struct{ name, uid string }

// TestMergeEventList_UnionsAndDedupsByUID verifies events are merged across
// must-gathers and deduplicated by uid, keeping the most recent capture's copy.
func TestMergeEventList_UnionsAndDedupsByUID(t *testing.T) {
	newRoot := t.TempDir()
	oldRoot := t.TempDir()
	// Shared uid E1 (new copy named ev-new, old copy ev-old), plus distinct events.
	writeEvents(t, newRoot, "ns1", []ev{{"ev-new", "E1"}, {"only-new", "E2"}})
	writeEvents(t, oldRoot, "ns1", []ev{{"ev-old", "E1"}, {"only-old", "E3"}})

	list, ageRoot := mergeEventList([]string{newRoot, oldRoot}, newRoot, "ns1", false)
	if ageRoot != newRoot {
		t.Fatalf("expected age root %s, got %s", newRoot, ageRoot)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 merged events (E1,E2,E3), got %d", len(list.Items))
	}
	names := map[string]bool{}
	for _, e := range list.Items {
		names[e.Name] = true
	}
	for _, want := range []string{"ev-new", "only-new", "only-old"} {
		if !names[want] {
			t.Errorf("expected merged events to contain %q; got %v", want, names)
		}
	}
	if names["ev-old"] {
		t.Errorf("expected older copy ev-old of uid E1 to be deduped out; got %v", names)
	}
}

// TestMergeEventList_FallsBackWhenMostRecentEmpty verifies that when the most
// recent capture has no events the merge still surfaces the older capture's.
func TestMergeEventList_FallsBackWhenMostRecentEmpty(t *testing.T) {
	newRoot := t.TempDir()
	oldRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(newRoot, "namespaces", "ns1", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeEvents(t, oldRoot, "ns1", []ev{{"ev-a", "E1"}, {"ev-b", "E2"}})

	list, _ := mergeEventList([]string{newRoot, oldRoot}, newRoot, "ns1", false)
	if len(list.Items) != 2 {
		t.Fatalf("expected the older capture's 2 events, got %d", len(list.Items))
	}
}
