// Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
			vars.MustGatherRootPath = "../../testdata/"
			vars.Namespace = tt.namespace
			validateArgs(tt.rtype)
			handleOutput(&stdout, &stderr)
			if !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("Got: %v \n", stderr.String())
				t.Errorf("Want: %v \n", tt.want)
			}
			vars.GetArgs = make(map[string]map[string]struct{})
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

	saved := vars.MustGatherRootPath
	t.Cleanup(func() { vars.MustGatherRootPath = saved })
	vars.MustGatherRootPath = root

	if err := getClusterScopedResources("clusterversions", "config.openshift.io", nil); err == nil {
		t.Fatalf("expected error from corrupt yaml, got nil")
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
	savedArgs := vars.GetArgs
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedPath
		vars.GetArgs = savedArgs
		GetCmd.SetArgs(nil)
		GetCmd.SetOut(nil)
		GetCmd.SetErr(nil)
	})
	vars.MustGatherRootPath = root
	vars.GetArgs = make(map[string]map[string]struct{})

	GetCmd.SetArgs([]string{"clusterversions"})
	GetCmd.SetOut(new(bytes.Buffer))
	GetCmd.SetErr(new(bytes.Buffer))

	if err := GetCmd.Execute(); err == nil {
		t.Fatalf("expected GetCmd.Execute to return an error from the corrupt fixture, got nil")
	}
}

func TestHandleObject_ReturnsErrorOnBadCustomColumns(t *testing.T) {
	savedOutput := vars.OutputStringVar
	savedNs := vars.Namespace
	savedSel := vars.LabelSelectorStringVar
	t.Cleanup(func() {
		vars.OutputStringVar = savedOutput
		vars.Namespace = savedNs
		vars.LabelSelectorStringVar = savedSel
	})
	vars.OutputStringVar = "custom-columns=BAD"
	vars.Namespace = ""
	vars.LabelSelectorStringVar = ""

	obj := unstructured.Unstructured{}
	obj.SetAPIVersion("v1")
	obj.SetKind("ConfigMap")
	obj.SetName("test")

	if err := handleObject(obj); err == nil {
		t.Fatalf("expected handleObject to return error for malformed custom-columns spec, got nil")
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
	savedArgs := vars.GetArgs
	savedOutput := vars.OutputStringVar
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedPath
		vars.GetArgs = savedArgs
		vars.OutputStringVar = savedOutput
		GetCmd.SetArgs(nil)
		GetCmd.SetOut(nil)
		GetCmd.SetErr(nil)
	})
	vars.MustGatherRootPath = root
	vars.GetArgs = make(map[string]map[string]struct{})

	GetCmd.SetArgs([]string{"clusterversions", "-o", "custom-columns=BAD"})
	GetCmd.SetOut(new(bytes.Buffer))
	GetCmd.SetErr(new(bytes.Buffer))

	if err := GetCmd.Execute(); err == nil {
		t.Fatalf("expected GetCmd.Execute to surface the CustomColumnsTable error, got nil")
	}
}
