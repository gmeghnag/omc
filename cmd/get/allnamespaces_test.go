package get

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeAggregatedPods writes namespaces/<ns>/core/pods.yaml listing the given
// pod names, all Running 1/1.
func writeAggregatedPods(t *testing.T, root, namespace string, names ...string) {
	t.Helper()
	coreDir := filepath.Join(root, "namespaces", namespace, "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("apiVersion: v1\nkind: List\nitems:\n")
	for _, name := range names {
		b.WriteString("- apiVersion: v1\n  kind: Pod\n  metadata:\n    name: " + name + "\n    namespace: " + namespace + "\n")
		b.WriteString("  spec:\n    containers:\n    - name: c\n      image: img\n")
		b.WriteString("  status:\n    phase: Running\n    containerStatuses:\n    - name: c\n      ready: true\n      restartCount: 0\n      state:\n        running: {}\n")
	}
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// podNameColumn returns the NAME column (2nd field, after NAMESPACE) of each
// data row in an "-A" pod table.
func podNameColumn(output string) []string {
	var names []string
	for _, line := range strings.Split(output, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] == "NAMESPACE" {
			continue
		}
		names = append(names, f[1])
	}
	return names
}

// TestGetPods_AllNamespaces_GlobalSort verifies --sort-by orders results across
// all namespaces, not independently within each namespace.
func TestGetPods_AllNamespaces_GlobalSort(t *testing.T) {
	root := t.TempDir()
	writeAggregatedPods(t, root, "aaa", "m-pod", "z-pod")
	writeAggregatedPods(t, root, "bbb", "a-pod", "n-pod")

	var out, errOut bytes.Buffer
	opts := Options{RootPath: root, AllNamespaces: true, SortBy: "{.metadata.name}"}
	if err := Run(&out, &errOut, opts, []string{"pods"}); err != nil {
		t.Fatalf("Run: %v (stderr: %s)", err, errOut.String())
	}
	got := podNameColumn(out.String())
	want := []string{"a-pod", "m-pod", "n-pod", "z-pod"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("global sort across namespaces:\n got: %v\nwant: %v\noutput:\n%s", got, want, out.String())
	}
}

// TestGetPods_AllNamespaces_ParallelDeterministic reads many namespaces (which
// exercises the bounded parallel reader) and asserts the output is complete and
// identical across runs. Combined with -race this guards the concurrency.
func TestGetPods_AllNamespaces_ParallelDeterministic(t *testing.T) {
	root := t.TempDir()
	const n = 25
	var expected []string
	for i := range n {
		ns := fmt.Sprintf("ns-%02d", i)
		pod := fmt.Sprintf("pod-%02d", i)
		writeAggregatedPods(t, root, ns, pod)
		expected = append(expected, pod)
	}
	sort.Strings(expected)

	var first string
	for run := range 5 {
		var out, errOut bytes.Buffer
		if err := Run(&out, &errOut, Options{RootPath: root, AllNamespaces: true}, []string{"pods"}); err != nil {
			t.Fatalf("run %d: %v (stderr: %s)", run, err, errOut.String())
		}
		names := podNameColumn(out.String())
		sort.Strings(names)
		if strings.Join(names, ",") != strings.Join(expected, ",") {
			t.Fatalf("run %d: missing/extra pods\n got: %v\nwant: %v", run, names, expected)
		}
		if run == 0 {
			first = out.String()
		} else if out.String() != first {
			t.Fatalf("run %d output differs from run 0 (non-deterministic):\n--- run0 ---\n%s\n--- run%d ---\n%s", run, first, run, out.String())
		}
	}
}
