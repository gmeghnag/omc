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
	vars.MustGatherRootPath = root
	t.Cleanup(func() { vars.MustGatherRootPath = oldRoot })

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
