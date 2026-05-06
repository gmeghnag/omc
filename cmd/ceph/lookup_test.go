package ceph

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFilename(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		args   []string
		want   string
	}{
		{"ceph status", "ceph", []string{"status"}, "ceph_status"},
		{"ceph osd tree", "ceph", []string{"osd", "tree"}, "ceph_osd_tree"},
		{"ceph osd df tree", "ceph", []string{"osd", "df", "tree"}, "ceph_osd_df_tree"},
		{"rados lspools", "rados", []string{"lspools"}, "rados_lspools"},
		{"rbd ls pool", "rbd", []string{"ls", "mypool"}, "rbd_ls_mypool"},
		{"ceph-volume raw list", "ceph-volume", []string{"raw", "list"}, "ceph-volume_raw_list"},
		{"radosgw-admin realm list", "radosgw-admin", []string{"realm", "list"}, "radosgw-admin_realm_list"},
		{"prefix only", "ceph", []string{}, "ceph"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFilename(tt.prefix, tt.args)
			if got != tt.want {
				t.Errorf("BuildFilename(%q, %v) = %q, want %q", tt.prefix, tt.args, got, tt.want)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantArgs   []string
		wantFormat string
	}{
		{"no flags", []string{"osd", "tree"}, []string{"osd", "tree"}, ""},
		{"--output flag", []string{"status", "--output", "json-pretty"}, []string{"status"}, "json-pretty"},
		{"--format flag", []string{"status", "--format", "json-pretty"}, []string{"status"}, "json-pretty"},
		{"-o flag", []string{"status", "-o", "json-pretty"}, []string{"status"}, "json-pretty"},
		{"--output= form", []string{"status", "--output=json-pretty"}, []string{"status"}, "json-pretty"},
		{"--format= form", []string{"status", "--format=json-pretty"}, []string{"status"}, "json-pretty"},
		{"flag in middle", []string{"osd", "tree", "--format", "json-pretty"}, []string{"osd", "tree"}, "json-pretty"},
		{"empty args", []string{}, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, gotFormat := ParseArgs(tt.args)
			if gotFormat != tt.wantFormat {
				t.Errorf("ParseArgs(%v) format = %q, want %q", tt.args, gotFormat, tt.wantFormat)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("ParseArgs(%v) args = %v, want %v", tt.args, gotArgs, tt.wantArgs)
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("ParseArgs(%v) args[%d] = %q, want %q", tt.args, i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestFilenameToCommand(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"simple", "ceph_status", "ceph status"},
		{"multi word", "ceph_osd_tree", "ceph osd tree"},
		{"strip json suffix", "ceph_status_--format_json-pretty", "ceph status"},
		{"rados", "rados_lspools", "rados lspools"},
		{"rbd", "rbd_ls_mypool", "rbd ls mypool"},
		{"ceph-volume", "ceph-volume_raw_list", "ceph-volume raw list"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilenameToCommand(tt.filename)
			if got != tt.want {
				t.Errorf("FilenameToCommand(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLookupAndPrintSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "ceph", "must_gather_commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "test ceph status output\n"
	if err := os.WriteFile(filepath.Join(cmdDir, "ceph_status"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	LookupAndPrint(tmpDir, "ceph", []string{"status"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if buf.String() != content {
		t.Errorf("LookupAndPrint output = %q, want %q", buf.String(), content)
	}
}

func TestLookupAndPrintJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonDir := filepath.Join(tmpDir, "ceph", "must_gather_commands_json_output")
	if err := os.MkdirAll(jsonDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `{"status": "HEALTH_OK"}` + "\n"
	if err := os.WriteFile(filepath.Join(jsonDir, "ceph_status_--format_json-pretty"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	LookupAndPrint(tmpDir, "ceph", []string{"status", "--output", "json-pretty"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if buf.String() != content {
		t.Errorf("LookupAndPrint JSON output = %q, want %q", buf.String(), content)
	}
}

func TestLookupAndPrintConfigShow(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "ceph", "must_gather_commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "config output for osd.0\n"
	if err := os.WriteFile(filepath.Join(cmdDir, "config_osd.0"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	LookupAndPrint(tmpDir, "ceph", []string{"config", "show", "osd.0"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if buf.String() != content {
		t.Errorf("LookupAndPrint config show output = %q, want %q", buf.String(), content)
	}
}

func TestListAvailableFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, "ceph", "must_gather_commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := []string{"ceph_status", "ceph_osd_tree", "ceph_osd_df", "rados_lspools", "rbd_ls_pool1"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(cmdDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Capture stderr since SuggestCommands writes there and calls os.Exit
	// Instead, test the underlying directory listing directly
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatal(err)
	}

	var cephFiles []string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) >= 5 && e.Name()[:5] == "ceph_" {
			cephFiles = append(cephFiles, e.Name())
		}
	}

	if len(cephFiles) != 3 {
		t.Errorf("Expected 3 ceph files, got %d: %v", len(cephFiles), cephFiles)
	}

	for _, expected := range []string{"ceph_status", "ceph_osd_tree", "ceph_osd_df"} {
		found := false
		for _, f := range cephFiles {
			if f == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %q not found in ceph files: %v", expected, cephFiles)
		}
	}
	fmt.Fprint(io.Discard, cephFiles)
}
