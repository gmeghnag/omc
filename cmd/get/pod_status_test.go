package get

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPodStatusShowsRunningWhenInitialized(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "namespaces/openshift-etcd/core")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Incomplete initContainerStatuses trigger Init:0/3 in the kubectl printer even
	// though the pod is fully initialized and running (common in must-gather).
	podList := []byte(`apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: etcd-test
    namespace: openshift-etcd
    labels:
      app: etcd
  spec:
    initContainers:
    - name: setup
      image: setup:latest
    - name: sidecar
      image: sidecar:latest
      restartPolicy: Always
    - name: cert
      image: cert:latest
    containers:
    - name: etcd
      image: etcd:latest
    - name: etcdctl
      image: etcdctl:latest
    - name: etcd-metrics
      image: metrics:latest
    - name: installer
      image: installer:latest
  status:
    phase: Running
    conditions:
    - type: Initialized
      status: "True"
    - type: Ready
      status: "True"
    initContainerStatuses:
    - name: setup
      ready: true
    - name: sidecar
      ready: true
      started: true
      state:
        running:
          startedAt: "2025-01-01T00:00:00Z"
    - name: cert
      ready: true
      state:
        terminated:
          exitCode: 0
          reason: Completed
    containerStatuses:
    - name: etcd
      ready: true
      state:
        running:
          startedAt: "2025-01-01T00:00:00Z"
    - name: etcdctl
      ready: true
      state:
        running:
          startedAt: "2025-01-01T00:00:00Z"
    - name: etcd-metrics
      ready: true
      state:
        running:
          startedAt: "2025-01-01T00:00:00Z"
    - name: installer
      ready: true
      state:
        running:
          startedAt: "2025-01-01T00:00:00Z"
`)
	if err := os.WriteFile(filepath.Join(dir, "pods.yaml"), podList, 0o644); err != nil {
		t.Fatal(err)
	}

	opts := newOptions()
	opts.RootPath = root
	opts.Namespace = "openshift-etcd"
	opts.LabelSelector = "app=etcd"
	validateArgs(&opts, []string{"pods"})

	var out bytes.Buffer
	if err := Run(&out, &bytes.Buffer{}, opts, []string{"pods"}); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if strings.Contains(output, "Init:") {
		t.Fatalf("expected Running status, got misleading init status:\n%s", output)
	}
	if !strings.Contains(output, "Running") {
		t.Fatalf("expected Running status in output:\n%s", output)
	}
}
