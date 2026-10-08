package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writePodsRoot writes an aggregated core/pods.yaml with two pods in namespace
// ns: a Running 1/1 pod and a Pending 0/1 pod.
func writePodsRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	coreDir := filepath.Join(root, "namespaces", "ns", "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	list := `apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: pod-running
    namespace: ns
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
- apiVersion: v1
  kind: Pod
  metadata:
    name: pod-pending
    namespace: ns
    creationTimestamp: "2023-11-02T06:00:00Z"
  spec:
    containers:
    - name: c
      image: img
  status:
    phase: Pending
    containerStatuses:
    - name: c
      ready: false
      restartCount: 0
      state:
        waiting: {}
`
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func fieldAfterName(t *testing.T, line string) []string {
	t.Helper()
	return strings.Fields(line)
}

// TestRun_TypedPodRenderingParity renders two pods through the typed path
// (which now resolves the internal type via the memoized RuntimeObjectForGVK)
// and asserts the kubectl-style columns and per-pod values are correct. This
// guards against regressions in type resolution and proves the cached type is
// applied correctly to every item, not just the first.
func TestRun_TypedPodRenderingParity(t *testing.T) {
	root := writePodsRoot(t)

	var out, errOut bytes.Buffer
	if err := Run(&out, &errOut, Options{RootPath: root, Namespace: "ns"}, []string{"pods"}); err != nil {
		t.Fatalf("Run: %v (stderr: %s)", err, errOut.String())
	}
	got := out.String()

	// Typed header from the kubectl pod printer.
	for _, col := range []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"} {
		if !strings.Contains(got, col) {
			t.Fatalf("missing typed column %q in output:\n%s", col, got)
		}
	}

	var runningLine, pendingLine string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "pod-running") {
			runningLine = line
		}
		if strings.HasPrefix(line, "pod-pending") {
			pendingLine = line
		}
	}
	if runningLine == "" || pendingLine == "" {
		t.Fatalf("expected both pod rows, got:\n%s", got)
	}

	// Columns: NAME READY STATUS RESTARTS AGE
	if f := fieldAfterName(t, runningLine); f[1] != "1/1" || f[2] != "Running" {
		t.Errorf("pod-running: want READY=1/1 STATUS=Running, got READY=%s STATUS=%s (%q)", f[1], f[2], runningLine)
	}
	if f := fieldAfterName(t, pendingLine); f[1] != "0/1" || f[2] != "Pending" {
		t.Errorf("pod-pending: want READY=0/1 STATUS=Pending, got READY=%s STATUS=%s (%q)", f[1], f[2], pendingLine)
	}
}
