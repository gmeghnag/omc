package get

import (
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// podWithRestarts builds a minimal Pod with one container. restartCount is an
// int64, the type numbers get when a must-gather YAML is decoded into
// Unstructured.
func podWithRestarts(name string, restartCount int64) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata":   map[string]interface{}{"name": name},
		"status": map[string]interface{}{
			"containerStatuses": []interface{}{
				map[string]interface{}{"name": "c", "restartCount": restartCount},
			},
		},
	}}
}

// pvWithCapacity builds a minimal PersistentVolume with the given storage
// capacity.
func pvWithCapacity(name, storage string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "PersistentVolume",
		"metadata":   map[string]interface{}{"name": name},
		"spec": map[string]interface{}{
			"capacity": map[string]interface{}{"storage": storage},
		},
	}}
}

func objectNames(list []unstructured.Unstructured) []string {
	var out []string
	for _, u := range list {
		out = append(out, u.GetName())
	}
	return out
}

// TestSortResources_NumericField uses the kubectl quick reference example
// "--sort-by='.status.containerStatuses[0].restartCount'". Numbers must be
// compared as numbers, not as text where "10" sorts before "9".
func TestSortResources_NumericField(t *testing.T) {
	list := []unstructured.Unstructured{
		podWithRestarts("restart-2", 2),
		podWithRestarts("restart-10", 10),
		podWithRestarts("restart-9", 9),
		podWithRestarts("restart-0", 0),
		podWithRestarts("restart-100", 100),
	}
	got := objectNames(sortResources(list, ".status.containerStatuses[0].restartCount"))
	want := []string{"restart-0", "restart-2", "restart-9", "restart-10", "restart-100"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestSortResources_QuantityField uses the kubectl quick reference example
// "--sort-by=.spec.capacity.storage". Like kubectl, values that parse as
// resource quantities are compared by amount.
func TestSortResources_QuantityField(t *testing.T) {
	list := []unstructured.Unstructured{
		pvWithCapacity("pv-10Gi", "10Gi"),
		pvWithCapacity("pv-500Mi", "500Mi"),
		pvWithCapacity("pv-1Ti", "1Ti"),
		pvWithCapacity("pv-2Gi", "2Gi"),
		pvWithCapacity("pv-100Gi", "100Gi"),
	}
	got := objectNames(sortResources(list, ".spec.capacity.storage"))
	want := []string{"pv-500Mi", "pv-2Gi", "pv-10Gi", "pv-100Gi", "pv-1Ti"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestSortResources_StringField checks that ordinary strings keep sorting as
// text, as kubectl does: "pod-10" sorts before "pod-9".
func TestSortResources_StringField(t *testing.T) {
	list := []unstructured.Unstructured{
		podWithRestarts("pod-9", 0),
		podWithRestarts("pod-10", 0),
		podWithRestarts("pod-1", 0),
	}
	got := objectNames(sortResources(list, ".metadata.name"))
	want := []string{"pod-1", "pod-10", "pod-9"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
