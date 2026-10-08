package logs

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveLogsRoot picks the most recent root that holds the pod's log tree
// and never falls back to an older capture once a newer one owns the pod.
func TestResolveLogsRoot(t *testing.T) {
	newRoot := t.TempDir()
	oldRoot := t.TempDir()
	// Pod lives (per-pod dir) only in the older capture.
	mkPodDir(t, oldRoot, "ns1", "p1")

	opts := Options{RootPath: newRoot, RootPaths: []string{newRoot, oldRoot}, Namespace: "ns1"}
	if got := resolveLogsRoot(opts, "p1"); got != oldRoot {
		t.Fatalf("expected logs to resolve to old root %s, got %s", oldRoot, got)
	}

	// Now the pod also exists in the newer capture: the newer one must win.
	mkPodDir(t, newRoot, "ns1", "p1")
	if got := resolveLogsRoot(opts, "p1"); got != newRoot {
		t.Fatalf("expected most recent root %s to win, got %s", newRoot, got)
	}

	// Single-must-gather context: always the one root, no resolution.
	single := Options{RootPath: newRoot, RootPaths: []string{newRoot}, Namespace: "ns1"}
	if got := resolveLogsRoot(single, "whatever"); got != newRoot {
		t.Fatalf("single-root context must return its root %s, got %s", newRoot, got)
	}
}

func mkPodDir(t *testing.T, root, ns, pod string) {
	t.Helper()
	dir := filepath.Join(root, "namespaces", ns, "pods", pod)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, pod+".yaml"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
}
