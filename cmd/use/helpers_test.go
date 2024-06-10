package use

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecompression(t *testing.T) {
	want := "testdata/must-gather.sample"
	tests := map[string]func(string, string) (string, error){
		"testdata/must-gather.zip":    ExtractZip,
		"testdata/must-gather.tar":    ExtractTar,
		"testdata/must-gather.tar.gz": ExtractTarGz,
		"testdata/must-gather.tar.xz": extractTarXZ,
	}

	for path, f := range tests {
		mgRootDir, err := f(path, "testdata")
		if err != nil {
			t.Error(err)
		}
		defer clearTestFiles(t)

		if filepath.Clean(want) != filepath.Clean(mgRootDir) {
			t.Errorf("expected %q, got %q", want, mgRootDir)
		}
		clearTestFiles(t)
	}
}

func clearTestFiles(t *testing.T) {
	t.Helper()

	err := os.RemoveAll("testdata/must-gather.sample")
	if err != nil {
		t.Fatal(err)
	}
}
