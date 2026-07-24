// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"
)

func TestHandleEmptyWideOutput(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		rtype     []string
		resources *types.UnstructuredList
		want      string
	}{
		{
			name:      "single cluster scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single cluster scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single namespaced scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakenamespacescopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single namespaced scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakenamespacescopedresources.operator.openshift.io found in default namespace.\n",
		},
		{
			name:      "cluster and namespaced scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "cluster and namespaced scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io found.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			opts := newOptions()
			opts.RootPath = "../../testdata/"
			opts.Namespace = tt.namespace
			validateArgs(&opts, tt.rtype)
			newState(&opts).handleOutput(&stdout, &stderr)
			if !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("Got: %v \n", stderr.String())
				t.Errorf("Want: %v \n", tt.want)
			}
		})
	}
}

func TestGetClusterScopedResources_ReturnsErrorOnCorruptYAML(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), []byte("{ unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := newOptions()
	opts.RootPath = root
	if err := getClusterScopedResources(newState(&opts), "clusterversions", "config.openshift.io", nil); err == nil {
		t.Fatalf("expected error from corrupt yaml, got nil")
	}
}

func TestGetClusterScopedResources_AcceptsListInPerResourceDir(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io", "clusterversions")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: config.openshift.io/v1
  kind: ClusterVersion
  metadata:
    name: version-a
- apiVersion: config.openshift.io/v1
  kind: ClusterVersion
  metadata:
    name: version-b
`)
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), fixture, 0o644); err != nil {
		t.Fatal(err)
	}

	opts := newOptions()
	opts.RootPath = root
	s := newState(&opts)
	if err := getClusterScopedResources(s, "clusterversions", "config.openshift.io", nil); err != nil {
		t.Fatalf("getClusterScopedResources: %v", err)
	}
	var out, errOut bytes.Buffer
	if err := s.handleOutput(&out, &errOut); err != nil {
		t.Fatalf("handleOutput: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "version-a") || !strings.Contains(got, "version-b") {
		t.Fatalf("expected both items in output, got:\n%s", got)
	}
}

func TestGetCmd_PropagatesErrorThroughCobra(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), []byte("{ unterminated"), 0o644); err != nil {
		t.Fatal(err)
	}

	savedPath := vars.MustGatherRootPath
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedPath
		GetCmd.SetArgs(nil)
		GetCmd.SetOut(nil)
		GetCmd.SetErr(nil)
	})
	vars.MustGatherRootPath = root

	GetCmd.SetArgs([]string{"clusterversions"})
	GetCmd.SetOut(new(bytes.Buffer))
	GetCmd.SetErr(new(bytes.Buffer))

	if err := GetCmd.Execute(); err == nil {
		t.Fatalf("expected GetCmd.Execute to return an error from the corrupt fixture, got nil")
	}
}

func TestHandleObject_ReturnsErrorOnBadCustomColumns(t *testing.T) {
	savedOutput := vars.OutputStringVar
	t.Cleanup(func() { vars.OutputStringVar = savedOutput })
	// tablegenerator still reads vars.OutputStringVar, so bridge it for this
	// unit test that bypasses Run.
	vars.OutputStringVar = "custom-columns=BAD"

	opts := newOptions()
	opts.Output = "custom-columns=BAD"

	obj := unstructured.Unstructured{}
	obj.SetAPIVersion("v1")
	obj.SetKind("ConfigMap")
	obj.SetName("test")

	if err := newState(&opts).handleObject(obj); err == nil {
		t.Fatalf("expected handleObject to return error for malformed custom-columns spec, got nil")
	}
}

func TestGetState_IsolatedAcrossInvocations(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: config.openshift.io/v1
  kind: ClusterVersion
  metadata:
    name: version
  status:
    desired:
      version: "4.17.11"
`)
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), fixture, 0o644); err != nil {
		t.Fatal(err)
	}

	run := func() string {
		opts := newOptions()
		opts.RootPath = root
		opts.GetArgs = map[string]map[string]struct{}{
			"clusterversions.config.openshift.io": {},
		}
		s := newState(&opts)
		if err := getClusterScopedResources(s, "clusterversions", "config.openshift.io", nil); err != nil {
			t.Fatalf("getClusterScopedResources: %v", err)
		}
		var out, errOut bytes.Buffer
		if err := s.handleOutput(&out, &errOut); err != nil {
			t.Fatalf("handleOutput: %v", err)
		}
		return out.String()
	}

	first := run()
	if first == "" {
		t.Fatalf("expected non-empty output from fixture, got empty")
	}
	second := run()
	if first != second {
		t.Fatalf("state leaks across invocations.\nfirst:\n%q\nsecond:\n%q", first, second)
	}
}

func TestGetCmd_PropagatesHandleObjectError(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: config.openshift.io/v1
  kind: ClusterVersion
  metadata:
    name: version
  status:
    desired:
      version: "4.17.11"
`)
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), fixture, 0o644); err != nil {
		t.Fatal(err)
	}

	savedPath := vars.MustGatherRootPath
	savedOutput := vars.OutputStringVar
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedPath
		vars.OutputStringVar = savedOutput
		GetCmd.SetArgs(nil)
		GetCmd.SetOut(nil)
		GetCmd.SetErr(nil)
	})
	vars.MustGatherRootPath = root

	GetCmd.SetArgs([]string{"clusterversions", "-o", "custom-columns=BAD"})
	GetCmd.SetOut(new(bytes.Buffer))
	GetCmd.SetErr(new(bytes.Buffer))

	if err := GetCmd.Execute(); err == nil {
		t.Fatalf("expected GetCmd.Execute to surface the CustomColumnsTable error, got nil")
	}
}

// TestRun_LibraryAndCobraParity checks that calling Run directly with an
// Options produces the same result as GetCmd.Execute() with the equivalent
// flags.
func TestRun_LibraryAndCobraParity(t *testing.T) {
	root := t.TempDir()
	rdir := filepath.Join(root, "cluster-scoped-resources", "config.openshift.io")
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: config.openshift.io/v1
  kind: ClusterVersion
  metadata:
    name: version
  status:
    desired:
      version: "4.17.11"
`)
	if err := os.WriteFile(filepath.Join(rdir, "clusterversions.yaml"), fixture, 0o644); err != nil {
		t.Fatal(err)
	}

	savedPath := vars.MustGatherRootPath
	savedOutput := vars.OutputStringVar
	savedNs := vars.Namespace
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedPath
		vars.OutputStringVar = savedOutput
		vars.Namespace = savedNs
		GetCmd.SetArgs(nil)
		GetCmd.SetOut(nil)
		GetCmd.SetErr(nil)
	})
	vars.MustGatherRootPath = root

	var libOut, libErr bytes.Buffer
	opts := newOptions()
	opts.RootPath = root
	opts.Output = "yaml"
	if err := Run(&libOut, &libErr, opts, []string{"clusterversions"}); err != nil {
		t.Fatalf("library Run: %v", err)
	}

	var cobraOut, cobraErr bytes.Buffer
	GetCmd.SetArgs([]string{"clusterversions", "-o", "yaml"})
	GetCmd.SetOut(&cobraOut)
	GetCmd.SetErr(&cobraErr)
	if err := GetCmd.Execute(); err != nil {
		t.Fatalf("GetCmd.Execute: %v", err)
	}

	if libOut.String() != cobraOut.String() {
		t.Fatalf("stdout drift between library and cobra paths\nlibrary:\n%s\ncobra:\n%s", libOut.String(), cobraOut.String())
	}
	if libOut.Len() == 0 {
		t.Fatalf("expected non-empty output from fixture")
	}
}

// TestRun_ConcurrentOptionsIsolation runs two goroutines calling Run with
// different Options and checks each call sees only its own options. The
// tablegenerator functions used to read vars globals, so concurrent calls
// raced on fields like vars.ShowKind and vars.Namespace.
func TestRun_ConcurrentOptionsIsolation(t *testing.T) {
	t.Parallel()

	// Build a shared fixture with two namespaces and one pod in each.
	root := t.TempDir()
	for _, ns := range []string{"ns-a", "ns-b"} {
		dir := filepath.Join(root, "namespaces", ns, "core")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		pods := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: pod-` + ns + `
    namespace: ` + ns + `
  spec: {}
  status: {}
`)
		if err := os.WriteFile(filepath.Join(dir, "pods.yaml"), pods, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	const iterations = 50
	type result struct {
		out string
		ns  string
	}
	results := make([]result, iterations*2)
	var wg sync.WaitGroup
	for i := range iterations {
		wg.Add(2)
		go func() {
			defer wg.Done()
			opts := newOptions()
			opts.RootPath = root
			opts.Namespace = "ns-a"
			var out, errOut bytes.Buffer
			if err := Run(&out, &errOut, opts, []string{"pods"}); err != nil {
				t.Errorf("iteration %d ns-a: %v", i, err)
				return
			}
			results[i*2] = result{out: out.String(), ns: "ns-a"}
		}()
		go func() {
			defer wg.Done()
			opts := newOptions()
			opts.RootPath = root
			opts.Namespace = "ns-b"
			var out, errOut bytes.Buffer
			if err := Run(&out, &errOut, opts, []string{"pods"}); err != nil {
				t.Errorf("iteration %d ns-b: %v", i, err)
				return
			}
			results[i*2+1] = result{out: out.String(), ns: "ns-b"}
		}()
		wg.Wait()
	}

	for i, r := range results {
		if r.ns == "ns-a" && !strings.Contains(r.out, "pod-ns-a") {
			t.Errorf("result %d (ns-a): expected pod-ns-a in output, got:\n%s", i, r.out)
		}
		if r.ns == "ns-a" && strings.Contains(r.out, "pod-ns-b") {
			t.Errorf("result %d (ns-a): found pod-ns-b in ns-a output, got:\n%s", i, r.out)
		}
		if r.ns == "ns-b" && !strings.Contains(r.out, "pod-ns-b") {
			t.Errorf("result %d (ns-b): expected pod-ns-b in output, got:\n%s", i, r.out)
		}
		if r.ns == "ns-b" && strings.Contains(r.out, "pod-ns-a") {
			t.Errorf("result %d (ns-b): found pod-ns-a in ns-b output, got:\n%s", i, r.out)
		}
	}
}
