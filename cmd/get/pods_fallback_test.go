package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writePerPodFiles writes pods under namespaces/<ns>/pods/<pod>/<pod>.yaml.
func writePerPodFiles(t *testing.T, root, namespace string, names ...string) {
	t.Helper()
	for _, name := range names {
		dir := filepath.Join(root, "namespaces", namespace, "pods", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		pod := `apiVersion: v1
kind: Pod
metadata:
  name: ` + name + `
  namespace: ` + namespace + `
  creationTimestamp: "2023-11-02T06:00:00Z"
spec:
  containers:
  - name: c
    image: img
status:
  phase: Running
  containerStatuses:
  - name: c
    ready: true
    restartCount: 0
    state:
      running: {}
`
		if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(pod), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// writeFlatPerPodFiles writes pods under namespaces/<ns>/core/pods/<pod>.yaml,
// the layout produced by e.g. the cluster-logging-operator must-gather image,
// where each pod directory instead holds its container logs
// (core/pods/<pod>/<container>/logs).
func writeFlatPerPodFiles(t *testing.T, root, namespace string, names ...string) {
	t.Helper()
	podsDir := filepath.Join(root, "namespaces", namespace, "core", "pods")
	for _, name := range names {
		if err := os.MkdirAll(podsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		pod := `apiVersion: v1
kind: Pod
metadata:
  name: ` + name + `
  namespace: ` + namespace + `
  creationTimestamp: "2023-11-02T06:00:00Z"
spec:
  containers:
  - name: c
    image: img
status:
  phase: Running
  containerStatuses:
  - name: c
    ready: true
    restartCount: 0
    state:
      running: {}
`
		if err := os.WriteFile(filepath.Join(podsDir, name+".yaml"), []byte(pod), 0o644); err != nil {
			t.Fatal(err)
		}
		// The log directory sharing the pod's name must not be mistaken for a
		// manifest, nor abort the read.
		logsDir := filepath.Join(podsDir, name, "c", "logs")
		if err := os.MkdirAll(logsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(logsDir, "current.log"), []byte("log line\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runGetPods(t *testing.T, root string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	if err := Run(&out, &errOut, Options{RootPath: root, Namespace: "ns"}, []string{"pods"}); err != nil {
		t.Fatalf("Run: %v (stderr: %s)", err, errOut.String())
	}
	return out.String()
}

// TestGetPods_PerPodFallbackWhenAggregatedAbsent covers a must-gather that only
// has the per-pod layout and no aggregated core/pods.yaml. Before the fix this
// returned "No resources pods found".
func TestGetPods_PerPodFallbackWhenAggregatedAbsent(t *testing.T) {
	root := t.TempDir()
	writePerPodFiles(t, root, "ns", "pod-a", "pod-b")

	got := runGetPods(t, root)
	for _, want := range []string{"pod-a", "pod-b"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

// TestGetPods_PerPodFallbackWhenAggregatedEmpty covers an empty (0-byte)
// aggregated core/pods.yaml alongside the per-pod layout. An empty file does
// not raise an unmarshal error, so the fallback is driven by the item count.
func TestGetPods_PerPodFallbackWhenAggregatedEmpty(t *testing.T) {
	root := t.TempDir()
	writePerPodFiles(t, root, "ns", "pod-a", "pod-b")
	coreDir := filepath.Join(root, "namespaces", "ns", "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got := runGetPods(t, root)
	for _, want := range []string{"pod-a", "pod-b"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

// TestGetPods_FlatPerPodLayout covers a must-gather with no aggregated
// core/pods.yaml whose pods live only in core/pods/<pod>.yaml (the layout used
// by the cluster-logging-operator image). This regressed in 4b38ba32, which
// only consulted the pods/<pod>/<pod>.yaml layout, so `get pods` reported
// "No resources pods found" for a namespace full of pods.
func TestGetPods_FlatPerPodLayout(t *testing.T) {
	root := t.TempDir()
	writeFlatPerPodFiles(t, root, "ns", "pod-a", "pod-b")

	got := runGetPods(t, root)
	for _, want := range []string{"pod-a", "pod-b"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "No resources") {
		t.Errorf("pods must be found in the flat core/pods layout, got:\n%s", got)
	}
}

// TestGetPods_FlatPerPodLayoutWhenAggregatedEmpty covers the flat layout next to
// an empty aggregated core/pods.yaml placeholder.
func TestGetPods_FlatPerPodLayoutWhenAggregatedEmpty(t *testing.T) {
	root := t.TempDir()
	writeFlatPerPodFiles(t, root, "ns", "pod-a")
	coreDir := filepath.Join(root, "namespaces", "ns", "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	got := runGetPods(t, root)
	if !strings.Contains(got, "pod-a") {
		t.Errorf("expected pod-a in output, got:\n%s", got)
	}
}

// TestGetPods_FlatLayoutIgnoresLogDirs ensures a pod directory holding only
// container logs contributes no pod and does not fail the read.
func TestGetPods_FlatLayoutIgnoresLogDirs(t *testing.T) {
	root := t.TempDir()
	writeFlatPerPodFiles(t, root, "ns", "pod-a")
	// A directory with no matching <pod>.yaml, as for a pod whose manifest was
	// not collected but whose logs were.
	if err := os.MkdirAll(filepath.Join(root, "namespaces", "ns", "core", "pods", "logs-only", "c", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := runGetPods(t, root)
	if !strings.Contains(got, "pod-a") {
		t.Fatalf("expected pod-a in output, got:\n%s", got)
	}
	if strings.Contains(got, "logs-only") {
		t.Errorf("a log-only directory must not render as a pod, got:\n%s", got)
	}
}

// TestGetPods_NestedPreferredOverFlat ensures the two per-pod layouts are not
// merged when both exist: the historical nested layout wins.
func TestGetPods_NestedPreferredOverFlat(t *testing.T) {
	root := t.TempDir()
	writePerPodFiles(t, root, "ns", "pod-nested")
	writeFlatPerPodFiles(t, root, "ns", "pod-flat")

	got := runGetPods(t, root)
	if !strings.Contains(got, "pod-nested") {
		t.Fatalf("expected pod-nested in output, got:\n%s", got)
	}
	if strings.Contains(got, "pod-flat") {
		t.Errorf("the flat layout must not be read when the nested one has pods, got:\n%s", got)
	}
}

// TestGetPods_AggregatedPreferredOverPerPod ensures a populated aggregated file
// is used as-is and the per-pod directory is not consulted (no duplication).
func TestGetPods_AggregatedPreferredOverPerPod(t *testing.T) {
	root := writePodsRoot(t) // aggregated core/pods.yaml with pod-running, pod-pending
	// Add a stray per-pod file that must be ignored when the aggregated list is present.
	writePerPodFiles(t, root, "ns", "pod-stray")

	got := runGetPods(t, root)
	if !strings.Contains(got, "pod-running") || !strings.Contains(got, "pod-pending") {
		t.Fatalf("expected aggregated pods, got:\n%s", got)
	}
	if strings.Contains(got, "pod-stray") {
		t.Errorf("per-pod directory must not be read when aggregated list is present, got:\n%s", got)
	}
}

// TestGetPods_AggregatedPreferredOverFlat is the same guard for the flat layout.
func TestGetPods_AggregatedPreferredOverFlat(t *testing.T) {
	root := writePodsRoot(t) // aggregated core/pods.yaml with pod-running, pod-pending
	writeFlatPerPodFiles(t, root, "ns", "pod-stray")

	got := runGetPods(t, root)
	if !strings.Contains(got, "pod-running") || !strings.Contains(got, "pod-pending") {
		t.Fatalf("expected aggregated pods, got:\n%s", got)
	}
	if strings.Contains(got, "pod-stray") {
		t.Errorf("flat per-pod layout must not be read when the aggregated list is present, got:\n%s", got)
	}
}

// TestGetNamespaced_PerFileCustomResource covers the refactored else-branch:
// a namespaced custom resource stored one file per object under
// namespaces/<ns>/<group>/<plural>/<name>.yaml.
func TestGetNamespaced_PerFileCustomResource(t *testing.T) {
	root := t.TempDir()
	group := "widgets.example.com"
	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.` + group + `
spec:
  group: ` + group + `
  names:
    kind: Widget
    plural: widgets
    singular: widget
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
`
	if err := os.WriteFile(filepath.Join(crdsDir, "widgets."+group+".yaml"), []byte(crd), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	wdir := filepath.Join(root, "namespaces", "ns", group, "widgets")
	if err := os.MkdirAll(wdir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"w1", "w2"} {
		obj := "apiVersion: " + group + "/v1\nkind: Widget\nmetadata:\n  name: " + name + "\n  namespace: ns\n"
		if err := os.WriteFile(filepath.Join(wdir, name+".yaml"), []byte(obj), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var out, errOut bytes.Buffer
	if err := Run(&out, &errOut, Options{RootPath: root, Namespace: "ns"}, []string{"widgets"}); err != nil {
		t.Fatalf("Run: %v (stderr: %s)", err, errOut.String())
	}
	got := out.String()
	for _, want := range []string{"w1", "w2"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected per-file CR %q in output, got:\n%s", want, got)
		}
	}
}
