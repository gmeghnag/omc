package use

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/vars"
)

// TestMustGatherInfo_EmptyInfrastructureItems guards against the index-out-of
// -range panic that occurred when infrastructures.yaml existed but carried no
// items (Items[0] was read unconditionally).
func TestMustGatherInfo_EmptyInfrastructureItems(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Present file, but empty items list.
	infra := "apiVersion: config.openshift.io/v1\nkind: InfrastructureList\nitems: []\n"
	if err := os.WriteFile(filepath.Join(dir, "infrastructures.yaml"), []byte(infra), 0o644); err != nil {
		t.Fatal(err)
	}

	oldRoot := vars.MustGatherRootPath
	oldRoots := vars.MustGatherRootPaths
	vars.MustGatherRootPath = root
	vars.MustGatherRootPaths = nil
	t.Cleanup(func() {
		vars.MustGatherRootPath = oldRoot
		vars.MustGatherRootPaths = oldRoots
	})

	// Silence stdout from the function.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	done := make(chan struct{})
	go func() { io.Copy(io.Discard, r); close(done) }()

	// Must not panic.
	MustGatherInfo()

	w.Close()
	os.Stdout = oldStdout
	<-done
}

// TestTimestampFromFile covers the multi-must-gather timestamp fix: the file is
// read whether it sits inside the root or next to it, a two-line file yields a
// range, a shorter one is INCOMPLETE, and an absent file is reported missing.
func TestTimestampFromFile(t *testing.T) {
	dir := t.TempDir()

	full := filepath.Join(dir, "full")
	if err := os.WriteFile(full, []byte(
		"2026-09-25 15:36:33.598482 +0200 CEST m=+0.217853001\n"+
			"2026-09-25 15:37:00.502472 +0200 CEST m=+27.121769126\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ts, ok := timestampFromFile(full); !ok || ts != "2026-09-25 15:36:33 - 2026-09-25 15:37:00" {
		t.Fatalf("full timestamp: ok=%v ts=%q", ok, ts)
	}

	one := filepath.Join(dir, "one")
	if err := os.WriteFile(one, []byte("2026-09-25 15:36:33.598482 +0200 CEST m=+0.2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ts, ok := timestampFromFile(one); !ok || ts != "INCOMPLETE" {
		t.Fatalf("single-line timestamp: ok=%v ts=%q (want INCOMPLETE)", ok, ts)
	}

	if ts, ok := timestampFromFile(filepath.Join(dir, "nope")); ok || ts != "" {
		t.Fatalf("missing timestamp: ok=%v ts=%q (want false/empty)", ok, ts)
	}
}
