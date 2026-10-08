package mustgather

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// writeRoot creates a must-gather root at base/<name>/inner (the inner dir holds
// namespaces/) with a sibling timestamp file carrying the given first line. It
// returns the root path (base/<name>/inner).
func writeRoot(t *testing.T, base, name, timestampLine string) string {
	t.Helper()
	top := filepath.Join(base, name)
	root := filepath.Join(top, "inner")
	if err := os.MkdirAll(filepath.Join(root, "namespaces"), 0o755); err != nil {
		t.Fatal(err)
	}
	if timestampLine != "" {
		if err := os.WriteFile(filepath.Join(top, "timestamp"), []byte(timestampLine+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func ts(line string) string { return line + " m=+0.000000001" }

func TestDiscoverRootsOrdersMostRecentFirst(t *testing.T) {
	base := t.TempDir()
	oldRoot := writeRoot(t, base, "mg-old", ts("2023-01-01 00:00:00.000000000 +0000 UTC"))
	newRoot := writeRoot(t, base, "mg-new", ts("2024-06-01 12:00:00.000000000 +0000 UTC"))

	roots, err := DiscoverRoots(base)
	if err != nil {
		t.Fatalf("DiscoverRoots: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d: %+v", len(roots), roots)
	}
	if roots[0].Path != newRoot || roots[1].Path != oldRoot {
		t.Fatalf("roots not most-recent-first:\n got  %s then %s\n want %s then %s",
			roots[0].Path, roots[1].Path, newRoot, oldRoot)
	}
}

func TestDiscoverRootsSingleMustGather(t *testing.T) {
	base := t.TempDir()
	root := writeRoot(t, base, "mg", ts("2024-01-01 00:00:00.000000000 +0000 UTC"))
	roots, err := DiscoverRoots(base)
	if err != nil {
		t.Fatalf("DiscoverRoots: %v", err)
	}
	if len(roots) != 1 || roots[0].Path != root {
		t.Fatalf("expected single root %s, got %+v", root, roots)
	}
}

// TestDiscoverRootsStopsAtRoot verifies namespace subdirectories inside a
// must-gather are not mistaken for separate roots.
func TestDiscoverRootsStopsAtRoot(t *testing.T) {
	base := t.TempDir()
	root := writeRoot(t, base, "mg", ts("2024-01-01 00:00:00.000000000 +0000 UTC"))
	// Add a nested namespace tree that itself contains no markers.
	if err := os.MkdirAll(filepath.Join(root, "namespaces", "openshift-foo", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	roots, err := DiscoverRoots(base)
	if err != nil {
		t.Fatalf("DiscoverRoots: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected exactly 1 root, got %d: %+v", len(roots), roots)
	}
}

// TestDiscoverRootsFallsBackToModTime ranks roots without a timestamp file by
// directory mod time.
func TestDiscoverRootsFallsBackToModTime(t *testing.T) {
	base := t.TempDir()
	older := writeRoot(t, base, "mg-a", "")
	newer := writeRoot(t, base, "mg-b", "")
	// Make newer's root clearly newer.
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()
	if err := os.Chtimes(older, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newer, recent, recent); err != nil {
		t.Fatal(err)
	}
	roots, err := DiscoverRoots(base)
	if err != nil {
		t.Fatalf("DiscoverRoots: %v", err)
	}
	if len(roots) != 2 || roots[0].Path != newer {
		t.Fatalf("expected %s first by mod time, got %+v", newer, roots)
	}
}

func pod(name, uid string) unstructured.Unstructured {
	obj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": name,
		},
	}
	if uid != "" {
		obj["metadata"].(map[string]interface{})["uid"] = uid
	}
	return unstructured.Unstructured{Object: obj}
}

func TestDedupByUIDKeepsFirstPerUID(t *testing.T) {
	// Ordered most-recent-first: the newest copy of uid "a" must win.
	items := []unstructured.Unstructured{
		pod("p-a", "a"), // newest a
		pod("p-b", "b"),
		pod("p-a-old", "a"), // older a, must be dropped
	}
	out := DedupByUID(items)
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d: %+v", len(out), names(out))
	}
	if out[0].GetName() != "p-a" || out[1].GetName() != "p-b" {
		t.Fatalf("most-recent copy not kept / order changed: %v", names(out))
	}
}

func TestDedupByUIDKeepsAllWithoutUID(t *testing.T) {
	items := []unstructured.Unstructured{
		pod("x", ""),
		pod("y", ""),
		pod("z", "z"),
		pod("z-dup", "z"),
	}
	out := DedupByUID(items)
	// Both uid-less items kept, plus one of the z's.
	if len(out) != 3 {
		t.Fatalf("expected 3 items (2 uid-less + 1 deduped), got %d: %v", len(out), names(out))
	}
}

func names(items []unstructured.Unstructured) []string {
	var n []string
	for _, i := range items {
		n = append(n, i.GetName())
	}
	return n
}

func TestResolveRoot(t *testing.T) {
	base := t.TempDir()
	newRoot := filepath.Join(base, "new")
	oldRoot := filepath.Join(base, "old")
	for _, r := range []string{newRoot, oldRoot} {
		if err := os.MkdirAll(filepath.Join(r, "namespaces", "ns"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// The pod lives only in the older root.
	podRel := "namespaces/ns/pods/p/p.yaml"
	if err := os.MkdirAll(filepath.Join(oldRoot, "namespaces/ns/pods/p"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldRoot, podRel), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	roots := []string{newRoot, oldRoot} // most-recent-first

	got, ok := ResolveRoot(roots, podRel)
	if !ok || got != oldRoot {
		t.Fatalf("expected to resolve to old root %s (ok=true), got %s ok=%v", oldRoot, got, ok)
	}

	// Not present anywhere: returns most recent, ok=false.
	got, ok = ResolveRoot(roots, "namespaces/ns/pods/missing/missing.yaml")
	if ok || got != newRoot {
		t.Fatalf("expected fallback to most recent %s ok=false, got %s ok=%v", newRoot, got, ok)
	}

	// Present in both: first (most recent) wins.
	if err := os.MkdirAll(filepath.Join(newRoot, "namespaces/ns/pods/p"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newRoot, podRel), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok = ResolveRoot(roots, podRel)
	if !ok || got != newRoot {
		t.Fatalf("expected most recent %s to win, got %s ok=%v", newRoot, got, ok)
	}
}
