package stern

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gmeghnag/omc/vars"
)

// testdata is the shared must-gather fixture used across omc's cmd
// packages (see cmd/logs/logreader_test.go, cmd/get/get_test.go). It
// contains two pods in "test-namespace" (test-pod with a single container,
// test-pod-2 with two containers) plus a third pod in "test-namespace-2",
// so a single fixture exercises multi-pod and multi-namespace matching.
const testdata = "../../testdata/"

func TestRun_MatchesPodsByRegexp(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata
	vars.Namespace = "test-namespace"

	var stdout, stderr bytes.Buffer
	opts := Options{Container: ".*", Tail: -1}
	if err := Run(&stdout, &stderr, opts, []string{"^test-pod"}); err != nil {
		t.Fatalf("Run: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "test-pod test-container") {
		t.Fatalf("expected test-pod/test-container log lines, got:\n%s", out)
	}
	if !strings.Contains(out, "test-pod-2 app test-pod-2 app log line") {
		t.Fatalf("expected test-pod-2/app log line, got:\n%s", out)
	}
	if !strings.Contains(out, "test-pod-2 sidecar test-pod-2 sidecar log line") {
		t.Fatalf("expected test-pod-2/sidecar log line, got:\n%s", out)
	}
}

func TestRun_AllNamespaces(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata

	var stdout, stderr bytes.Buffer
	opts := Options{Container: ".*", Tail: -1, AllNamespace: true}
	if err := Run(&stdout, &stderr, opts, []string{"^test-pod"}); err != nil {
		t.Fatalf("Run: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "test-pod test-container") {
		t.Fatalf("expected test-pod log line from test-namespace, got:\n%s", out)
	}
	if !strings.Contains(out, "test-pod-3 test-container-3 test-pod-3 log line") {
		t.Fatalf("expected test-pod-3 log line from test-namespace-2, got:\n%s", out)
	}
}

func TestRun_ContainerFilter(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata
	vars.Namespace = "test-namespace"

	var stdout, stderr bytes.Buffer
	opts := Options{Container: "^app$", Tail: -1}
	if err := Run(&stdout, &stderr, opts, []string{"^test-pod-2$"}); err != nil {
		t.Fatalf("Run: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "test-pod-2 app test-pod-2 app log line") {
		t.Fatalf("expected test-pod-2/app log line, got:\n%s", out)
	}
	if strings.Contains(out, "sidecar log line") {
		t.Fatalf("did not expect sidecar log line, got:\n%s", out)
	}
}

// TestSternCommand_QueryMatchesSubstringAnywhere confirms that, like
// upstream stern, POD_QUERY is matched unanchored (regexp.MatchString),
// i.e. it matches the pattern anywhere in the pod name, not only as a
// prefix. "another-app" contains "t" and "r" but does not start with
// either, so `omc stern t` must still match it alongside "test-pod" and
// "test-pod-2" (which do start with "t").
func TestSternCommand_QueryMatchesSubstringAnywhere(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata
	vars.Namespace = "test-namespace"

	var stdout, stderr bytes.Buffer
	Stern.SetOut(&stdout)
	Stern.SetErr(&stderr)
	Stern.SetArgs([]string{"t"})
	t.Cleanup(func() {
		Stern.SetArgs(nil)
		Stern.SetOut(nil)
		Stern.SetErr(nil)
	})

	if err := Stern.Execute(); err != nil {
		t.Fatalf("Stern.Execute: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "test-pod test-container") {
		t.Fatalf("expected test-pod to match query \"t\", got:\n%s", out)
	}
	if !strings.Contains(out, "test-pod-2 app") {
		t.Fatalf("expected test-pod-2 to match query \"t\", got:\n%s", out)
	}
	if !strings.Contains(out, "another-app main another-app main log line") {
		t.Fatalf("expected another-app to match query \"t\" as a substring (not just a prefix), got:\n%s", out)
	}
}

// TestRun_QueryRMatchesPodsContainingRAnywhere directly exercises the
// reported scenario: `omc stern r` must match every pod containing "r"
// anywhere in its name (here, only "another-app", which has "r" as its
// last letter), not pods that merely start with "r".
func TestRun_QueryRMatchesPodsContainingRAnywhere(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata
	vars.Namespace = "test-namespace"

	var stdout, stderr bytes.Buffer
	opts := Options{Container: ".*", Tail: -1}
	if err := Run(&stdout, &stderr, opts, []string{"r"}); err != nil {
		t.Fatalf("Run: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "another-app main another-app main log line") {
		t.Fatalf("expected another-app (contains \"r\", does not start with it) to match query \"r\", got:\n%s", out)
	}
	if strings.Contains(out, "test-pod ") || strings.Contains(out, "test-pod-2 ") {
		t.Fatalf("did not expect test-pod/test-pod-2 (no \"r\" in name) to match query \"r\", got:\n%s", out)
	}
}

func TestRun_NoMatchReturnsError(t *testing.T) {
	restoreSternState(t)
	vars.MustGatherRootPath = testdata
	vars.Namespace = "test-namespace"

	var stdout, stderr bytes.Buffer
	opts := Options{Container: ".*", Tail: -1}
	err := Run(&stdout, &stderr, opts, []string{"^nomatch$"})
	if err == nil {
		t.Fatalf("expected no-match error, got nil")
	}
	if !strings.Contains(err.Error(), "no pods found matching") {
		t.Fatalf("expected no-match error, got %v", err)
	}
}

func TestPrefixWriter(t *testing.T) {
	var buf bytes.Buffer
	pw := newPrefixWriter(&buf, "pod", "container")

	if _, err := pw.Write([]byte("hello ")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := pw.Write([]byte("world\nfoo")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := pw.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	want := "pod container hello world\npod container foo\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestPrefixWriter_FlushNoopWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	pw := newPrefixWriter(&buf, "pod", "container")
	if _, err := pw.Write([]byte("complete line\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := pw.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	want := "pod container complete line\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func restoreSternState(t *testing.T) {
	t.Helper()
	savedRoot := vars.MustGatherRootPath
	savedNamespace := vars.Namespace
	savedAllNamespace := vars.AllNamespaceBoolVar
	t.Cleanup(func() {
		vars.MustGatherRootPath = savedRoot
		vars.Namespace = savedNamespace
		vars.AllNamespaceBoolVar = savedAllNamespace
	})
}
