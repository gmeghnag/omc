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
package logs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gmeghnag/omc/vars"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestLogReaderReadReturnsWriterError(t *testing.T) {
	logReader := &LogReader{
		dirname: testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/",
		files:   &[]string{"current.log"},
		filter:  nil,
		tail:    -1,
	}

	err := logReader.Read(failingWriter{})
	if err == nil {
		t.Fatalf("expected writer error, got nil")
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("expected write failure in error, got %v", err)
	}
}

func TestFilterCatLogsReturnsErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		err := FilterCatLogs(filepath.Join(t.TempDir(), "missing.log"), []string{"info"})
		if err == nil {
			t.Fatalf("expected missing file error, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected missing file error, got %v", err)
		}
	})

	t.Run("malformed cri log", func(t *testing.T) {
		logPath := filepath.Join(t.TempDir(), "current.log")
		if err := os.WriteFile(logPath, []byte("not-a-cri-line\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := FilterCatLogs(logPath, []string{"info"})
		if err == nil {
			t.Fatalf("expected malformed log error, got nil")
		}
		if !strings.Contains(err.Error(), "timestamp is not found") {
			t.Fatalf("expected timestamp parse error, got %v", err)
		}
	})
}

func TestLogsPodsReturnsErrors(t *testing.T) {
	t.Run("missing pod", func(t *testing.T) {
		root := writePodsListFixture(t, podListYAML("other-pod", "test-container"))

		err := logsPods(io.Discard, root, "test-namespace", "test-pod", "", false, false, false, nil, false, -1)
		if err == nil {
			t.Fatalf("expected missing pod error, got nil")
		}
		if !strings.Contains(err.Error(), "pods test-pod not found") {
			t.Fatalf("expected missing pod error, got %v", err)
		}
	})

	t.Run("invalid container", func(t *testing.T) {
		root := writePodsListFixture(t, podListYAML("test-pod", "test-container"))

		err := logsPods(io.Discard, root, "test-namespace", "test-pod", "missing-container", false, false, false, nil, false, -1)
		if err == nil {
			t.Fatalf("expected invalid container error, got nil")
		}
		if !strings.Contains(err.Error(), "container missing-container is not valid for pod test-pod") {
			t.Fatalf("expected invalid container error, got %v", err)
		}
	})

	t.Run("corrupt pods list", func(t *testing.T) {
		root := writePodsListFixture(t, "{ unterminated")

		err := logsPods(io.Discard, root, "test-namespace", "test-pod", "", false, false, false, nil, false, -1)
		if err == nil {
			t.Fatalf("expected corrupt pods list error, got nil")
		}
		if !strings.Contains(err.Error(), "error unmarshaling") {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})

	t.Run("corrupt fallback pod", func(t *testing.T) {
		root := t.TempDir()
		podDir := filepath.Join(root, "namespaces", "test-namespace", "pods", "test-pod")
		if err := os.MkdirAll(podDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(podDir, "test-pod.yaml"), []byte("{ unterminated"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := logsPods(io.Discard, root, "test-namespace", "test-pod", "", false, false, false, nil, false, -1)
		if err == nil {
			t.Fatalf("expected corrupt fallback pod error, got nil")
		}
		if !strings.Contains(err.Error(), "error unmarshaling") {
			t.Fatalf("expected unmarshal error, got %v", err)
		}
	})
}

func TestLogsCommandReturnsErrors(t *testing.T) {
	t.Run("invalid argument count", func(t *testing.T) {
		root := writeLogsRoot(t)
		restoreLogsCommandState(t)
		vars.MustGatherRootPath = root

		var stdout, stderr bytes.Buffer
		Logs.SetOut(&stdout)
		Logs.SetErr(&stderr)
		Logs.SetArgs([]string{})

		err := Logs.Execute()
		if err == nil {
			t.Fatalf("expected argument error, got nil")
		}
		if !strings.Contains(err.Error(), "POD or TYPE/NAME is a required argument") {
			t.Fatalf("expected argument error, got %v", err)
		}
	})

	t.Run("container flag conflicts with inline container", func(t *testing.T) {
		root := writeLogsRoot(t)
		restoreLogsCommandState(t)
		vars.MustGatherRootPath = root

		var stdout, stderr bytes.Buffer
		Logs.SetOut(&stdout)
		Logs.SetErr(&stderr)
		Logs.SetArgs([]string{"-c", "flag-container", "test-pod", "inline-container"})

		err := Logs.Execute()
		if err == nil {
			t.Fatalf("expected container conflict error, got nil")
		}
		if !strings.Contains(err.Error(), "only one of -c or an inline [CONTAINER] arg is allowed") {
			t.Fatalf("expected container conflict error, got %v", err)
		}
	})
}

func writeLogsRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "namespaces"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func writePodsListFixture(t *testing.T, podsYAML string) string {
	t.Helper()
	root := t.TempDir()
	coreDir := filepath.Join(root, "namespaces", "test-namespace", "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "pods.yaml"), []byte(podsYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func podListYAML(podName, containerName string) string {
	return `apiVersion: v1
kind: PodList
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: ` + podName + `
  spec:
    containers:
    - name: ` + containerName + `
`
}

// TestRun_LibraryAndCobraParity checks that calling Run directly with an
// Options produces the same result as Logs.Execute() with the equivalent
// flags.
func TestRun_LibraryAndCobraParity(t *testing.T) {
	root := writePodsListFixture(t, podListYAML("test-pod", "test-container"))
	logDir := filepath.Join(root, "namespaces", "test-namespace", "pods", "test-pod", "test-container", "test-container", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "current.log"), []byte("hello world\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	restoreLogsCommandState(t)
	vars.MustGatherRootPath = root
	vars.Namespace = "test-namespace"

	var libOut, libErr bytes.Buffer
	opts := Options{RootPath: root, Namespace: "test-namespace", Container: "test-container", Tail: -1}
	if err := Run(&libOut, &libErr, opts, []string{"test-pod"}); err != nil {
		t.Fatalf("library Run: %v", err)
	}

	var cobraOut, cobraErr bytes.Buffer
	Logs.SetOut(&cobraOut)
	Logs.SetErr(&cobraErr)
	Logs.SetArgs([]string{"-c", "test-container", "test-pod"})
	if err := Logs.Execute(); err != nil {
		t.Fatalf("Logs.Execute: %v", err)
	}

	if libOut.String() != cobraOut.String() {
		t.Fatalf("stdout drift between library and cobra paths\nlibrary:\n%s\ncobra:\n%s", libOut.String(), cobraOut.String())
	}
	if libOut.Len() == 0 {
		t.Fatalf("expected non-empty output from fixture")
	}
}

func restoreLogsCommandState(t *testing.T) {
	t.Helper()
	savedPath := vars.MustGatherRootPath
	savedNamespace := vars.Namespace
	savedContainer := containerFlag
	savedPrevious := previousFlag
	savedRotated := rotatedFlag
	savedAllContainers := allContainersFlag
	savedInsecureLogs := insecureFlag
	savedTail := tailFlag
	savedLogLevel := LogLevel

	t.Cleanup(func() {
		Logs.SetArgs(nil)
		Logs.SetOut(nil)
		Logs.SetErr(nil)
		_ = Logs.PersistentFlags().Set("container", savedContainer)
		_ = Logs.PersistentFlags().Set("previous", strconv.FormatBool(savedPrevious))
		_ = Logs.PersistentFlags().Set("rotated", strconv.FormatBool(savedRotated))
		_ = Logs.PersistentFlags().Set("all-containers", strconv.FormatBool(savedAllContainers))
		_ = Logs.PersistentFlags().Set("insecure", strconv.FormatBool(savedInsecureLogs))
		_ = Logs.PersistentFlags().Set("tail", strconv.FormatInt(savedTail, 10))
		vars.MustGatherRootPath = savedPath
		vars.Namespace = savedNamespace
		LogLevel = savedLogLevel
	})
}
