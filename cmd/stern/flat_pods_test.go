package stern

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gmeghnag/omc/vars"
)

// writeFlatPod writes a pod in the flat layout used by e.g. the
// cluster-logging-operator must-gather image: <ns>/core/pods/<pod>.yaml, with
// the pod's container logs in a sibling directory of the same name.
func writeFlatPod(t *testing.T, root, namespace, name, container, logLine string) {
	t.Helper()
	podsDir := filepath.Join(root, "namespaces", namespace, "core", "pods")
	if err := os.MkdirAll(podsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: ` + name + `
  namespace: ` + namespace + `
spec:
  containers:
  - name: ` + container + `
    image: img
status:
  phase: Running
`
	if err := os.WriteFile(filepath.Join(podsDir, name+".yaml"), []byte(pod), 0o644); err != nil {
		t.Fatal(err)
	}
	logsDir := filepath.Join(podsDir, name, container, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "current.log"), []byte(logLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRun_FlatPodLayout covers a must-gather with no aggregated core/pods.yaml
// whose pods live only in core/pods/<pod>.yaml. stern used to find no pods in
// such a capture and fail with "no pods found matching", because it only looked
// at core/pods.yaml and the nested pods/<pod>/<pod>.yaml layout.
func TestRun_FlatPodLayout(t *testing.T) {
	restoreSternState(t)
	root := t.TempDir()
	writeFlatPod(t, root, "logging", "collector-abc", "collector", "flat layout log line")
	vars.MustGatherRootPath = root
	vars.Namespace = "logging"

	var stdout, stderr bytes.Buffer
	opts := Options{Container: ".*", Tail: -1}
	if err := Run(&stdout, &stderr, opts, []string{"^collector-abc$"}); err != nil {
		t.Fatalf("Run: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "collector-abc collector") {
		t.Fatalf("expected collector-abc/collector log lines, got:\n%s", out)
	}
	if !strings.Contains(out, "flat layout log line") {
		t.Errorf("expected the captured log line, got:\n%s", out)
	}
}

// TestListPods_FlatLayout checks the pod listing that backs the matching.
func TestListPods_FlatLayout(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, root, "logging", "collector-abc", "collector", "line")

	pods, err := listPods(filepath.Join(root, "namespaces", "logging"))
	if err != nil {
		t.Fatalf("listPods: %v", err)
	}
	if len(pods) != 1 || pods[0].Name != "collector-abc" {
		t.Fatalf("expected [collector-abc], got %v", pods)
	}
}

// TestListPods_IgnoresLogOnlyDirs ensures a directory holding only container
// logs is not reported as a pod.
func TestListPods_IgnoresLogOnlyDirs(t *testing.T) {
	root := t.TempDir()
	writeFlatPod(t, root, "logging", "collector-abc", "collector", "line")
	logsOnly := filepath.Join(root, "namespaces", "logging", "core", "pods", "logs-only", "c", "logs")
	if err := os.MkdirAll(logsOnly, 0o755); err != nil {
		t.Fatal(err)
	}

	pods, err := listPods(filepath.Join(root, "namespaces", "logging"))
	if err != nil {
		t.Fatalf("listPods: %v", err)
	}
	for _, pod := range pods {
		if pod.Name == "logs-only" {
			t.Errorf("a log-only directory must not be reported as a pod")
		}
	}
}
