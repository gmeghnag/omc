package logs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFlatPod writes a pod in the flat layout used by e.g. the
// cluster-logging-operator must-gather image: <ns>/core/pods/<pod>.yaml with the
// pod's logs in a sibling directory of the same name.
func writeFlatPod(t *testing.T, root, ns, pod, container, logLine string) {
	t.Helper()
	podsDir := filepath.Join(root, "namespaces", ns, "core", "pods")
	if err := os.MkdirAll(podsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: ` + pod + `
  namespace: ` + ns + `
spec:
  containers:
  - name: ` + container + `
    image: img
`
	if err := os.WriteFile(filepath.Join(podsDir, pod+".yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	logsDir := filepath.Join(podsDir, pod, container, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "current.log"), []byte(logLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runLogsPod(t *testing.T, root, ns, pod string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	// Tail: -1 mirrors the CLI default (show all lines); the zero value means
	// "no lines", so it must be set explicitly.
	err := Run(&stdout, &stderr, Options{RootPath: root, Namespace: ns, Tail: -1}, []string{pod})
	return stdout.String(), err
}

// TestLogs_FlatPodLayout covers a must-gather whose pods live only in
// core/pods/<pod>.yaml (no aggregated core/pods.yaml).
func TestLogs_FlatPodLayout(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, root, "ns", "pod-a", "c", "flat log line")

	out, err := runLogsPod(t, root, "ns", "pod-a")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "flat log line") {
		t.Errorf("expected the flat layout log line, got:\n%s", out)
	}
}

// TestLogs_FlatPodLayoutWithEmptyAggregated covers an empty (0-byte)
// core/pods.yaml placeholder next to the flat layout. Reading the aggregated
// file directly made this fail with "pods <name> not found", because an empty
// file unmarshals to an empty list without error and so never triggered the
// per-pod fallback.
func TestLogs_FlatPodLayoutWithEmptyAggregated(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, root, "ns", "pod-a", "c", "flat log line")
	coreDir := filepath.Join(root, "namespaces", "ns", "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runLogsPod(t, root, "ns", "pod-a")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "flat log line") {
		t.Errorf("expected the flat layout log line, got:\n%s", out)
	}
}

// TestLogs_MissingPodIsAnError keeps the "not found" path honest now that the
// resolution goes through helpers.PodObjects.
func TestLogs_MissingPodIsAnError(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, root, "ns", "pod-a", "c", "line")

	_, err := runLogsPod(t, root, "ns", "nope")
	if err == nil {
		t.Fatal("expected an error for a pod that is not in the must-gather")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected a 'not found' error, got: %v", err)
	}
}
