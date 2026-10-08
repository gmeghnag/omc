package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTwoClusterScopedKinds builds a must-gather root with two cluster-scoped
// CRDs (widgets.<ga>, gadgets.<gb>) and one object of each, so a multi-resource
// get produces two distinct, non-empty blocks.
func writeTwoClusterScopedKinds(t *testing.T, ga, gb string) string {
	t.Helper()
	root := t.TempDir()
	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeKind := func(plural, singular, kind, group, item string) {
		crd := "apiVersion: apiextensions.k8s.io/v1\n" +
			"kind: CustomResourceDefinition\n" +
			"metadata:\n  name: " + plural + "." + group + "\n" +
			"spec:\n  group: " + group + "\n  names:\n    kind: " + kind + "\n    plural: " + plural + "\n    singular: " + singular + "\n  scope: Cluster\n" +
			"  versions:\n  - name: v1\n    served: true\n    storage: true\n    schema:\n      openAPIV3Schema:\n        type: object\n"
		if err := os.WriteFile(filepath.Join(crdsDir, plural+"."+group+".yaml"), []byte(crd), 0o644); err != nil {
			t.Fatal(err)
		}
		resDir := filepath.Join(root, "cluster-scoped-resources", group)
		if err := os.MkdirAll(resDir, 0o755); err != nil {
			t.Fatal(err)
		}
		list := "apiVersion: v1\nkind: List\nitems:\n- apiVersion: " + group + "/v1\n  kind: " + kind + "\n  metadata:\n    name: " + item + "\n"
		if err := os.WriteFile(filepath.Join(resDir, plural+".yaml"), []byte(list), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeKind("widgets", "widget", "Widget", ga, "widget-1")
	writeKind("gadgets", "gadget", "Gadget", gb, "gadget-1")
	return root
}

// TestRun_MultiResourceOutputOrderDeterministic asserts that "widgets,gadgets"
// always renders the widgets block before the gadgets block. Before the fix,
// Run ranged over the GetArgs map, so the block order was random.
func TestRun_MultiResourceOutputOrderDeterministic(t *testing.T) {
	root := writeTwoClusterScopedKinds(t, "a.example.com", "b.example.com")
	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	const runs = 40
	for i := range runs {
		var out, errOut bytes.Buffer
		if err := Run(&out, &errOut, Options{RootPath: root}, []string{"widgets,gadgets"}); err != nil {
			t.Fatalf("run %d: %v (stderr: %s)", i, err, errOut.String())
		}
		got := out.String()
		wIdx := strings.Index(got, "widget-1")
		gIdx := strings.Index(got, "gadget-1")
		if wIdx < 0 || gIdx < 0 {
			t.Fatalf("run %d: expected both blocks, got:\n%s", i, got)
		}
		if wIdx > gIdx {
			t.Fatalf("run %d: widgets must precede gadgets (requested order), got:\n%s", i, got)
		}
	}
}
