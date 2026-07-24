/*
Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// writeClusterScopedRoot builds a root with a cluster-scoped Widget CRD in group
// plus one Widget named itemName.
func writeClusterScopedRoot(t *testing.T, group, itemName string) string {
	t.Helper()
	root := t.TempDir()

	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.` + group + `
spec:
  group: ` + group + `
  names:
    kind: Widget
    plural: widgets
    singular: widget
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
`
	if err := os.WriteFile(filepath.Join(crdsDir, "widgets."+group+".yaml"), []byte(crd), 0o644); err != nil {
		t.Fatal(err)
	}

	resDir := filepath.Join(root, "cluster-scoped-resources", group)
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		t.Fatal(err)
	}
	list := `apiVersion: v1
kind: List
items:
- apiVersion: ` + group + `/v1
  kind: Widget
  metadata:
    name: ` + itemName + `
`
	if err := os.WriteFile(filepath.Join(resDir, "widgets.yaml"), []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// writeNamespacedRoot builds a root with a namespaced Widget CRD in group plus
// one Widget named itemName in the default namespace.
func writeNamespacedRoot(t *testing.T, group, itemName string) string {
	t.Helper()
	root := t.TempDir()

	crdsDir := filepath.Join(root, "cluster-scoped-resources", "apiextensions.k8s.io", "customresourcedefinitions")
	if err := os.MkdirAll(crdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.` + group + `
spec:
  group: ` + group + `
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
	if err := os.WriteFile(filepath.Join(crdsDir, "widgets."+group+".yaml"), []byte(crd), 0o644); err != nil {
		t.Fatal(err)
	}

	resDir := filepath.Join(root, "namespaces", "default", group)
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		t.Fatal(err)
	}
	list := `apiVersion: v1
kind: List
items:
- apiVersion: ` + group + `/v1
  kind: Widget
  metadata:
    name: ` + itemName + `
    namespace: default
`
	if err := os.WriteFile(filepath.Join(resDir, "widgets.yaml"), []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestRunConcurrentDistinctRootsResolveOwnCrd checks alias resolution is isolated
// per call. Both roots define "widget" with a different group and scope, so a
// shared lookup would resolve one call against the other bundle's CRD.
func TestRunConcurrentDistinctRootsResolveOwnCrd(t *testing.T) {
	rootA := writeClusterScopedRoot(t, "a.example.com", "widget-a")
	rootB := writeNamespacedRoot(t, "b.example.com", "widget-b")

	t.Cleanup(func() {
		crdCache.Lock()
		for _, r := range []string{rootA, rootB} {
			delete(crdCache.byRoot, r)
		}
		crdCache.Unlock()
	})

	var (
		wg         sync.WaitGroup
		outA, outB bytes.Buffer
		errA, errB bytes.Buffer
		runErrA    error
		runErrB    error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		runErrA = Run(&outA, &errA, Options{RootPath: rootA, Namespace: "default"}, []string{"widget"})
	}()
	go func() {
		defer wg.Done()
		runErrB = Run(&outB, &errB, Options{RootPath: rootB, Namespace: "default"}, []string{"widget"})
	}()
	wg.Wait()

	if runErrA != nil {
		t.Fatalf("root A Run: %v (stderr: %s)", runErrA, errA.String())
	}
	if runErrB != nil {
		t.Fatalf("root B Run: %v (stderr: %s)", runErrB, errB.String())
	}

	gotA := outA.String()
	if !strings.Contains(gotA, "widget-a") {
		t.Errorf("root A output should resolve its own cluster-scoped Widget, got:\n%s\nstderr:\n%s", gotA, errA.String())
	}
	if strings.Contains(gotA, "widget-b") {
		t.Errorf("root A output leaked root B's Widget, got:\n%s", gotA)
	}

	gotB := outB.String()
	if !strings.Contains(gotB, "widget-b") {
		t.Errorf("root B output should resolve its own namespaced Widget, got:\n%s\nstderr:\n%s", gotB, errB.String())
	}
	if strings.Contains(gotB, "widget-a") {
		t.Errorf("root B output leaked root A's Widget, got:\n%s", gotB)
	}
}

// TestRunConcurrentSameRoot runs many Runs against one root at once. They share
// the parsed CRD cache but each holds its own alias map, so the race detector
// should stay quiet.
func TestRunConcurrentSameRoot(t *testing.T) {
	root := writeClusterScopedRoot(t, "a.example.com", "widget-a")
	t.Cleanup(func() {
		crdCache.Lock()
		delete(crdCache.byRoot, root)
		crdCache.Unlock()
	})

	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	outs := make([]bytes.Buffer, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			var stderr bytes.Buffer
			errs[i] = Run(&outs[i], &stderr, Options{RootPath: root, Namespace: "default"}, []string{"widget"})
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Errorf("run %d: %v", i, errs[i])
			continue
		}
		if !strings.Contains(outs[i].String(), "widget-a") {
			t.Errorf("run %d did not resolve its Widget, got:\n%s", i, outs[i].String())
		}
	}
}
