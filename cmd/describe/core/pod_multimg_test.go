package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/pkg/mustgather"
)

// TestPodExistsInRoot covers both pod layouts: the per-pod directory and the
// aggregated core/pods.yaml list. Resolution must find a pod that lives only in
// an older capture, keeping describe consistent with get.
func TestPodExistsInRoot(t *testing.T) {
	root := t.TempDir()
	ns := filepath.Join(root, "namespaces", "ns1", "core")
	if err := os.MkdirAll(ns, 0o755); err != nil {
		t.Fatal(err)
	}
	// aggregated list holds "listed"; per-pod dir holds "perpod".
	if err := os.WriteFile(filepath.Join(ns, "pods.yaml"), []byte(`apiVersion: v1
kind: PodList
items:
- apiVersion: v1
  kind: Pod
  metadata: {name: listed, namespace: ns1, uid: U1}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	perpod := filepath.Join(root, "namespaces", "ns1", "pods", "perpod")
	if err := os.MkdirAll(perpod, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(perpod, "perpod.yaml"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !podExistsInRoot(root, "ns1", "listed") {
		t.Errorf("expected to find pod 'listed' via aggregated core/pods.yaml")
	}
	if !podExistsInRoot(root, "ns1", "perpod") {
		t.Errorf("expected to find pod 'perpod' via per-pod directory")
	}
	if podExistsInRoot(root, "ns1", "missing") {
		t.Errorf("did not expect to find absent pod 'missing'")
	}
}

// TestDescribePodResolution verifies the most recent capture containing a pod
// is chosen, including when the pod only exists in the aggregated list of the
// older capture.
func TestDescribePodResolution(t *testing.T) {
	newRoot := t.TempDir()
	oldRoot := t.TempDir()
	for _, r := range []string{newRoot, oldRoot} {
		if err := os.MkdirAll(filepath.Join(r, "namespaces", "ns1", "core"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// only-old exists only in the older capture's aggregated list.
	if err := os.WriteFile(filepath.Join(oldRoot, "namespaces", "ns1", "core", "pods.yaml"), []byte(`apiVersion: v1
kind: PodList
items:
- {apiVersion: v1, kind: Pod, metadata: {name: only-old, namespace: ns1, uid: U3}}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	roots := []string{newRoot, oldRoot}
	got, ok := mustgather.ResolveRootBy(roots, func(root string) bool {
		return podExistsInRoot(root, "ns1", "only-old")
	})
	if !ok || got != oldRoot {
		t.Fatalf("expected describe to resolve only-old to old root %s (ok=true), got %s ok=%v", oldRoot, got, ok)
	}
}
