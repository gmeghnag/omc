package get

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
)

func TestKindGroupNamespacedFromCrds_NilAliasCache(t *testing.T) {
	// External callers pass nil. The function must not panic on the writes it
	// does while scanning CRDs.
	root := t.TempDir()
	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	// No CRDs on disk, so this returns an error rather than panicking.
	if _, _, _, _, err := kindGroupNamespacedFromCrds("nonexistent", root, nil); err == nil {
		t.Error("expected error when no CRD matches, got nil")
	}
}

func TestKindGroupNamespacedFromCrds_HomedirFallback(t *testing.T) {
	// When the bundle has no matching CRD, ~/.omc/customresourcedefinitions is
	// an always-on fallback: a CRD present there resolves without any flag.
	root := t.TempDir() // bundle root with no CRD directory

	home := t.TempDir()
	t.Setenv("HOME", home)
	omcCrdsDir := filepath.Join(home, ".omc", "customresourcedefinitions")
	if err := os.MkdirAll(omcCrdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crdYAML := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gadgets.example.com
spec:
  group: example.com
  names:
    kind: Gadget
    plural: gadgets
    singular: gadget
  scope: Namespaced
`
	if err := os.WriteFile(filepath.Join(omcCrdsDir, "gadgets.example.com.yaml"), []byte(crdYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	plural, group, singular, namespaced, err := kindGroupNamespacedFromCrds("gadget", root, nil)
	if err != nil {
		t.Fatalf("expected homedir fallback to resolve the alias, got error: %v", err)
	}
	if plural != "gadgets" || group != "example.com" || singular != "gadget" || !namespaced {
		t.Errorf("unexpected resolution: plural=%q group=%q singular=%q namespaced=%v", plural, group, singular, namespaced)
	}
}

func TestKindGroupNamespacedFromCrds_Cache(t *testing.T) {
	// Two calls with the same MustGatherRootPath should return the same result.
	// After the first call the CRD file is made unreadable, so a second call
	// that still succeeds confirms the result came from the in-memory cache.
	root := t.TempDir()
	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crdYAML := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  names:
    kind: Widget
    plural: widgets
    singular: widget
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
`
	crdFile := filepath.Join(crdsDir, "widgets.example.com.yaml")
	if err := os.WriteFile(crdFile, []byte(crdYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
		os.Chmod(crdFile, 0o644)
	})

	plural1, group1, _, _, err := kindGroupNamespacedFromCrds("widget", root, nil)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Make the CRD file unreadable so a second disk read would fail.
	if err := os.Chmod(crdFile, 0o000); err != nil {
		t.Fatal(err)
	}
	// A fresh alias map means no fast path, so a second success proves
	// crdCache.byRoot served the parsed CRD without touching disk.
	plural2, group2, _, _, err := kindGroupNamespacedFromCrds("widget", root, nil)
	if err != nil {
		t.Fatalf("second call (expected cache hit): %v", err)
	}
	if plural1 != plural2 || group1 != group2 {
		t.Errorf("results differ: %q/%q vs %q/%q", plural1, group1, plural2, group2)
	}
}

func TestReadDirForResources(t *testing.T) {
	tests := []struct {
		name     string
		in       fstest.MapFS
		expected []string
	}{
		{
			name: "read correct resource files/dirs",
			in: fstest.MapFS{
				"resource-file-1.yaml":            {Data: []byte("abc")},
				"resource.yaml":                   {Data: []byte("abc")},
				"1.yaml":                          {Data: []byte("abc")},
				"resource.with.dot.filename.yaml": {Data: []byte("abc")},
				"resource-directory-name":         {Data: []byte("abc"), Mode: fs.ModeDir},
				"resource.directory.with.dot":     {Data: []byte("abc"), Mode: fs.ModeDir},
			},
			expected: []string{
				"resource-file-1.yaml",
				"resource.yaml",
				"1.yaml",
				"resource.with.dot.filename.yaml",
				"resource-directory-name",
				"resource.directory.with.dot",
			},
		},
		{
			name: "read only resource files/dirs matching the expected name convention",
			in: fstest.MapFS{
				"resource-file-1.yaml":             {Data: []byte("abc")},
				"._faulthy-resource-filename.yaml": {Data: []byte("abc")}, // e.g. AppleDouble encoded Macintosh file
				".resource-filename.yaml.swp":      {Data: []byte("abc")},
				"-resource-filename.yaml":          {Data: []byte("abc")},
			},
			expected: []string{"resource-file-1.yaml"},
		},
		{
			name: "read only resource files/dir with size > 0",
			in: fstest.MapFS{
				"resource-file-1.yaml":         {Data: []byte("abc")},
				"empty-resource-filename.yaml": {},
			},
			expected: []string{"resource-file-1.yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := readDirForResources(tc.in)
			if len(got) != len(tc.expected) {
				t.Errorf("Got: %v \n", got)
				t.Errorf("Want: %v \n", tc.expected)
			} else {
				gotNames := make([]string, 0)
				for _, dir := range got {
					gotNames = append(gotNames, dir.Name())
				}
				sort.Slice(gotNames, func(i, j int) bool {
					return gotNames[i] > gotNames[j]
				})
				sort.Slice(tc.expected, func(i, j int) bool {
					return tc.expected[i] > tc.expected[j]
				})
				if !reflect.DeepEqual(gotNames, tc.expected) {
					t.Error("Got:")
					for _, g := range gotNames {
						t.Error(g)
					}
					t.Error("Want:")
					for _, te := range tc.expected {
						t.Error(te)
					}
				}
			}
		})
	}
}
