package core

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFlatPod writes a pod in the flat layout used by e.g. the
// cluster-logging-operator must-gather image: <ns>/core/pods/<pod>.yaml, with
// the pod's container logs in a sibling directory of the same name.
func writeFlatPod(t *testing.T, nsPath, name string) {
	t.Helper()
	podsDir := filepath.Join(nsPath, "core", "pods")
	if err := os.MkdirAll(podsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: ` + name + `
  namespace: ns1
spec:
  containers:
  - name: c
    image: img
status:
  phase: Running
`
	if err := os.WriteFile(filepath.Join(podsDir, name+".yaml"), []byte(pod), 0o644); err != nil {
		t.Fatal(err)
	}
	logsDir := filepath.Join(podsDir, name, "c", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "current.log"), []byte("line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDescribePod_FlatLayout covers a must-gather with no aggregated
// core/pods.yaml whose pods live only in core/pods/<pod>.yaml. describe pod used
// to print nothing for such a capture, because it only looked at the aggregated
// list and the nested pods/<pod>/<pod>.yaml layout.
func TestDescribePod_FlatLayout(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, filepath.Join(root, "namespaces", "ns1"), "pod-flat")

	if !podExistsInRoot(root, "ns1", "pod-flat") {
		t.Errorf("expected to find pod 'pod-flat' via core/pods/<pod>.yaml")
	}
	describePod(root, "ns1", []string{"pod-flat"})
}
