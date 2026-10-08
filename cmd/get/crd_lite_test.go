package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseCRDLite checks the lightweight CRD decoder extracts exactly the
// fields omc needs (names, group, scope, versions + printer columns) and
// tolerates a large openAPIV3Schema it is meant to skip.
func TestParseCRDLite(t *testing.T) {
	// A schema big enough that skipping it is the whole point; its presence
	// must not affect the extracted fields.
	var schema strings.Builder
	schema.WriteString("      openAPIV3Schema:\n        type: object\n        properties:\n")
	for range 500 {
		schema.WriteString("          field")
		schema.WriteString(strings.Repeat("x", 3))
		schema.WriteString(": {type: string, description: \"" + strings.Repeat("y", 50) + "\"}\n")
	}

	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  scope: Namespaced
  names:
    kind: Widget
    plural: widgets
    singular: widget
    shortNames: [wi, wg]
  versions:
  - name: v1
    served: true
    storage: true
    additionalPrinterColumns:
    - name: Replicas
      type: integer
      jsonPath: .spec.replicas
    schema:
` + schema.String()

	got, err := parseCRDLite([]byte(crd))
	if err != nil {
		t.Fatalf("parseCRDLite: %v", err)
	}
	if got.Spec.Group != "example.com" {
		t.Errorf("group: got %q", got.Spec.Group)
	}
	if string(got.Spec.Scope) != "Namespaced" {
		t.Errorf("scope: got %q", got.Spec.Scope)
	}
	if got.Spec.Names.Kind != "Widget" || got.Spec.Names.Plural != "widgets" || got.Spec.Names.Singular != "widget" {
		t.Errorf("names: got %+v", got.Spec.Names)
	}
	if strings.Join(got.Spec.Names.ShortNames, ",") != "wi,wg" {
		t.Errorf("shortNames: got %v", got.Spec.Names.ShortNames)
	}
	if len(got.Spec.Versions) != 1 || got.Spec.Versions[0].Name != "v1" {
		t.Fatalf("versions: got %+v", got.Spec.Versions)
	}
	cols := got.Spec.Versions[0].AdditionalPrinterColumns
	if len(cols) != 1 || cols[0].Name != "Replicas" || cols[0].JSONPath != ".spec.replicas" || cols[0].Type != "integer" {
		t.Errorf("printer columns: got %+v", cols)
	}
}

// TestGetCustomResource_AdditionalPrinterColumns proves the lite-parsed printer
// columns flow through to rendering: a CRD-declared column is shown with the
// value pulled from the object.
func TestGetCustomResource_AdditionalPrinterColumns(t *testing.T) {
	root := t.TempDir()
	group := "example.com"
	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  scope: Namespaced
  names:
    kind: Widget
    plural: widgets
    singular: widget
  versions:
  - name: v1
    served: true
    storage: true
    additionalPrinterColumns:
    - name: Replicas
      type: integer
      jsonPath: .spec.replicas
    schema:
      openAPIV3Schema:
        type: object
`
	if err := os.WriteFile(filepath.Join(crdsDir, "widgets.example.com.yaml"), []byte(crd), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	resDir := filepath.Join(root, "namespaces", "ns", group)
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		t.Fatal(err)
	}
	list := `apiVersion: v1
kind: List
items:
- apiVersion: example.com/v1
  kind: Widget
  metadata:
    name: w1
    namespace: ns
  spec:
    replicas: 3
`
	if err := os.WriteFile(filepath.Join(resDir, "widgets.yaml"), []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := Run(&out, &errOut, Options{RootPath: root, Namespace: "ns"}, []string{"widgets"}); err != nil {
		t.Fatalf("Run: %v (stderr: %s)", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "REPLICAS") {
		t.Errorf("expected REPLICAS column header, got:\n%s", got)
	}
	var w1 string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "w1") {
			w1 = line
		}
	}
	if f := strings.Fields(w1); len(f) < 2 || f[1] != "3" {
		t.Errorf("expected w1 Replicas=3 from printer column, got %q (full:\n%s)", w1, got)
	}
}
