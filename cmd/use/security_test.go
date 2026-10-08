package use

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeExtractPath(t *testing.T) {
	dest := filepath.Clean("/tmp/omc-extract-dest")
	tests := []struct {
		name    string
		entry   string
		wantErr bool
	}{
		{"plain file", "mg/namespaces/pods.yaml", false},
		{"root file", "file.txt", false},
		{"parent escape", "../evil.txt", true},
		{"deep escape", "mg/../../../evil.txt", true},
		{"absolute is contained", "/etc/passwd", false}, // Join makes it relative to dest
		{"sneaky prefix sibling", "../omc-extract-dest-evil/x", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeExtractPath(dest, tt.entry)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for entry %q, got path %q", tt.entry, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for entry %q: %v", tt.entry, err)
			}
			if got != dest && !bytes.HasPrefix([]byte(got), []byte(dest+string(os.PathSeparator))) {
				t.Fatalf("path %q is not within dest %q", got, dest)
			}
		})
	}
}

// writeMaliciousTarGz builds a tar.gz with a benign entry and a traversal entry.
func writeMaliciousTarGz(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	write := func(name string, data []byte) {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	write("mg/namespaces/.placeholder", []byte(""))
	write("mg/../../../omc-escaped.txt", []byte("pwned"))
}

func writeMaliciousZip(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	for _, e := range []struct{ name, data string }{
		{"mg/namespaces/.placeholder", ""},
		{"mg/../../../omc-escaped.txt", "pwned"},
	} {
		w, err := zw.Create(e.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.data)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExtractTarGz_RejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	arch := filepath.Join(tmp, "evil.tar.gz")
	writeMaliciousTarGz(t, arch)
	dest := filepath.Join(tmp, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ExtractTarGz(arch, dest); err == nil {
		t.Fatalf("expected ExtractTarGz to reject a traversal entry")
	}
	// The escaped file must not exist anywhere outside dest.
	escaped := filepath.Join(filepath.Dir(filepath.Dir(dest)), "omc-escaped.txt")
	if _, err := os.Stat(escaped); err == nil {
		t.Fatalf("traversal entry escaped to %s", escaped)
	}
	if _, err := os.Stat(filepath.Join(tmp, "omc-escaped.txt")); err == nil {
		t.Fatalf("traversal entry escaped into tmp root")
	}
}

func TestExtractZip_RejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	arch := filepath.Join(tmp, "evil.zip")
	writeMaliciousZip(t, arch)
	dest := filepath.Join(tmp, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ExtractZip(arch, dest); err == nil {
		t.Fatalf("expected ExtractZip to reject a traversal entry")
	}
	if _, err := os.Stat(filepath.Join(tmp, "omc-escaped.txt")); err == nil {
		t.Fatalf("traversal entry escaped into tmp root")
	}
}

// TestExtractZip_ManyFilesNoFDLeak extracts an archive with many files. With the
// previous per-iteration defer this accumulated one open descriptor per entry;
// extracting cleanly documents that handles are released as it goes.
func TestExtractZip_ManyFilesNoFDLeak(t *testing.T) {
	tmp := t.TempDir()
	arch := filepath.Join(tmp, "many.zip")
	f, err := os.Create(arch)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	const n = 500
	for i := range n {
		w, err := zw.Create(fmt.Sprintf("mg/file-%04d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	dest := filepath.Join(tmp, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractZip(arch, dest); err != nil {
		t.Fatalf("ExtractZip of %d files failed: %v", n, err)
	}
	// Spot-check a few files landed.
	for _, i := range []int{0, n / 2, n - 1} {
		if _, err := os.Stat(filepath.Join(dest, "mg", fmt.Sprintf("file-%04d.txt", i))); err != nil {
			t.Fatalf("missing extracted file %d: %v", i, err)
		}
	}
}
