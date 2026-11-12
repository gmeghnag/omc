package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

func writePod(t *testing.T, path, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: ` + name + `
spec:
  containers:
  - name: c
    image: img
`
	if err := os.WriteFile(path, []byte(pod), 0o644); err != nil {
		t.Fatal(err)
	}
}

func podNames(t *testing.T, objects []PodObject) []string {
	t.Helper()
	var names []string
	for _, o := range objects {
		var pod struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}
		if err := yaml.Unmarshal(o.Data, &pod); err != nil {
			t.Fatalf("unmarshal %s: %v", o.Path, err)
		}
		names = append(names, pod.Metadata.Name)
	}
	return names
}

// TestPodObjects_Aggregated checks the classic layout: a List under
// <ns>/core/pods.yaml.
func TestPodObjects_Aggregated(t *testing.T) {
	nsPath := t.TempDir()
	aggregated := `apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata: {name: pod-a}
- apiVersion: v1
  kind: Pod
  metadata: {name: pod-b}
`
	if err := os.MkdirAll(filepath.Join(nsPath, "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nsPath, "core", "pods.yaml"), []byte(aggregated), 0o644); err != nil {
		t.Fatal(err)
	}

	objects, err := PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	names := podNames(t, objects)
	if len(names) != 2 || names[0] != "pod-a" || names[1] != "pod-b" {
		t.Errorf("expected [pod-a pod-b], got %v", names)
	}
}

// TestPodObjects_Nested checks the historical per-pod layout.
func TestPodObjects_Nested(t *testing.T) {
	nsPath := t.TempDir()
	writePod(t, filepath.Join(nsPath, "pods", "pod-a", "pod-a.yaml"), "pod-a")

	objects, err := PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if names := podNames(t, objects); len(names) != 1 || names[0] != "pod-a" {
		t.Errorf("expected [pod-a], got %v", names)
	}
}

// TestPodObjects_Flat covers the layout used by the cluster-logging-operator
// image: core/pods/<pod>.yaml with the pod's logs in a sibling directory of the
// same name. This is the case that regressed and made `get pods` print
// "No resources pods found" for a namespace full of pods.
func TestPodObjects_Flat(t *testing.T) {
	nsPath := t.TempDir()
	writePod(t, filepath.Join(nsPath, "core", "pods", "pod-a.yaml"), "pod-a")
	// Container logs living next to the manifests must be ignored.
	logsDir := filepath.Join(nsPath, "core", "pods", "pod-a", "c", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "current.log"), []byte("line\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	objects, err := PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if names := podNames(t, objects); len(names) != 1 || names[0] != "pod-a" {
		t.Errorf("expected [pod-a], got %v", names)
	}
}

// TestPodObjects_EmptyAggregatedFallsBack covers an empty aggregated
// placeholder, which unmarshals without error and must not mask the per-pod
// layouts.
func TestPodObjects_EmptyAggregatedFallsBack(t *testing.T) {
	nsPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(nsPath, "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nsPath, "core", "pods.yaml"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	writePod(t, filepath.Join(nsPath, "core", "pods", "pod-a.yaml"), "pod-a")

	objects, err := PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if names := podNames(t, objects); len(names) != 1 || names[0] != "pod-a" {
		t.Errorf("expected [pod-a], got %v", names)
	}
}

// TestPodObjects_LayoutPrecedence ensures layouts are not merged: the
// aggregated list wins over both per-pod layouts, and nested wins over flat.
func TestPodObjects_LayoutPrecedence(t *testing.T) {
	nsPath := t.TempDir()
	writePod(t, filepath.Join(nsPath, "pods", "nested", "nested.yaml"), "nested")
	writePod(t, filepath.Join(nsPath, "core", "pods", "flat.yaml"), "flat")

	objects, err := PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if names := podNames(t, objects); len(names) != 1 || names[0] != "nested" {
		t.Fatalf("expected [nested], got %v", names)
	}

	aggregated := "apiVersion: v1\nkind: List\nitems:\n- apiVersion: v1\n  kind: Pod\n  metadata: {name: agg}\n"
	if err := os.WriteFile(filepath.Join(nsPath, "core", "pods.yaml"), []byte(aggregated), 0o644); err != nil {
		t.Fatal(err)
	}
	objects, err = PodObjects(nsPath)
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if names := podNames(t, objects); len(names) != 1 || names[0] != "agg" {
		t.Errorf("expected [agg], got %v", names)
	}
}

// TestPodObjects_EmptyNamespace checks a namespace with no pods at all.
func TestPodObjects_EmptyNamespace(t *testing.T) {
	objects, err := PodObjects(t.TempDir())
	if err != nil {
		t.Fatalf("PodObjects: %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("expected no pods, got %d", len(objects))
	}
}

// TestPodExists covers resolution across every layout, including a pod that
// exists only in one of them while another layout is populated.
func TestPodExists(t *testing.T) {
	nsPath := t.TempDir()
	writePod(t, filepath.Join(nsPath, "pods", "nested", "nested.yaml"), "nested")
	writePod(t, filepath.Join(nsPath, "core", "pods", "flat.yaml"), "flat")
	aggregated := "apiVersion: v1\nkind: List\nitems:\n- apiVersion: v1\n  kind: Pod\n  metadata: {name: agg}\n"
	if err := os.WriteFile(filepath.Join(nsPath, "core", "pods.yaml"), []byte(aggregated), 0o644); err != nil {
		t.Fatal(err)
	}
	// A log-only directory is not a pod.
	if err := os.MkdirAll(filepath.Join(nsPath, "core", "pods", "logs-only", "c", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"nested", "flat", "agg"} {
		if !PodExists(nsPath, name) {
			t.Errorf("expected to find pod %q", name)
		}
	}
	for _, name := range []string{"missing", "logs-only"} {
		if PodExists(nsPath, name) {
			t.Errorf("did not expect to find pod %q", name)
		}
	}
}
