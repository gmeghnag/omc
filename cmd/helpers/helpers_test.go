package helpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestCachedModTime_MemoizesStat proves the Age/Last-Seen reference time is
// read from disk once and then served from cache, instead of being re-stat'd
// on every call (which happened once per rendered table row before the fix).
func TestCachedModTime_MemoizesStat(t *testing.T) {
	ResetModTimeCache()
	t.Cleanup(ResetModTimeCache)

	dir := t.TempDir()
	ns := filepath.Join(dir, "namespaces")
	if err := os.MkdirAll(ns, 0o755); err != nil {
		t.Fatal(err)
	}
	creation := metav1.NewTime(time.Now().Add(-time.Hour))

	first := TranslateTimestamp(dir, creation)
	if first == "<unknown>" {
		t.Fatalf("expected a duration, got %q", first)
	}

	// Move the directory's mod time far into the future. A non-cached
	// implementation would observe the new mod time and return a different age.
	future := time.Now().Add(72 * time.Hour)
	if err := os.Chtimes(ns, future, future); err != nil {
		t.Fatal(err)
	}
	if second := TranslateTimestamp(dir, creation); second != first {
		t.Fatalf("expected memoized result %q, got %q after mod-time change", first, second)
	}

	// After an explicit reset the fresh (future) mod time must be picked up,
	// confirming the first result really came from the cache.
	ResetModTimeCache()
	if third := TranslateTimestamp(dir, creation); third == first {
		t.Fatalf("expected a different age after reset, still got %q", third)
	}
}

// TestCachedModTime_UnknownWhenMissing keeps the fallback behaviour: a root
// without the reference paths yields the unknown marker.
func TestCachedModTime_UnknownWhenMissing(t *testing.T) {
	ResetModTimeCache()
	t.Cleanup(ResetModTimeCache)

	dir := t.TempDir() // no namespaces/ subdir
	creation := metav1.NewTime(time.Now().Add(-time.Hour))
	if got := TranslateTimestamp(dir, creation); got != "<unknown>" {
		t.Fatalf("expected <unknown> for a root without namespaces/, got %q", got)
	}
}
