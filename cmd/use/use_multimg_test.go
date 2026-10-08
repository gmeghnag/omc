package use

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"
)

func activePaths(t *testing.T, cfgPath string) (string, []string) {
	t.Helper()
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg types.Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, c := range cfg.Contexts {
		if c.Current == "*" {
			return c.Path, c.Paths
		}
	}
	t.Fatal("no active context")
	return "", nil
}

// mkRoot makes a minimal must-gather root (has namespaces/) at base/name.
func mkRoot(t *testing.T, base, name string) string {
	t.Helper()
	root := filepath.Join(base, name)
	if err := os.MkdirAll(filepath.Join(root, "namespaces", "ns1"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestUseContext_UpgradesExistingContextToGrouped reproduces the bug where
// re-using a path already present as a single-must-gather context did not
// persist the freshly discovered grouped roots.
func TestUseContext_UpgradesExistingContextToGrouped(t *testing.T) {
	t.Cleanup(func() { vars.MustGatherRootPath = ""; vars.MustGatherRootPaths = nil })
	tmp := t.TempDir()
	base := filepath.Join(tmp, "container")
	r0 := mkRoot(t, base, "mg-a")
	r1 := mkRoot(t, base, "mg-b")

	// Pre-existing single-must-gather context pointing at r0 (no paths).
	cfgPath := filepath.Join(tmp, "omc.json")
	initial := types.Config{Id: "x", Contexts: []types.Context{{Id: "x", Path: r0, Current: "*", Project: "default"}}}
	b, _ := json.MarshalIndent(initial, "", " ")
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatal(err)
	}

	// Re-use with a discovered grouping (most-recent-first [r0, r1]).
	if err := useContext(r0, cfgPath, "", []string{r0, r1}); err != nil {
		t.Fatalf("useContext: %v", err)
	}
	_, paths := activePaths(t, cfgPath)
	if len(paths) != 2 || paths[0] != r0 || paths[1] != r1 {
		t.Fatalf("expected grouped paths [r0 r1] persisted, got %v", paths)
	}

	// Re-using the same context as a single must-gather clears the grouping.
	if err := useContext(r0, cfgPath, "", nil); err != nil {
		t.Fatalf("useContext (single): %v", err)
	}
	if _, paths := activePaths(t, cfgPath); len(paths) != 0 {
		t.Fatalf("expected grouping cleared for single must-gather, got %v", paths)
	}
}
