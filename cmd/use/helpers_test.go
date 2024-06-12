package use

import (
	"os"
	"testing"
)

func TestDecompression(t *testing.T) {
	want := "testdata/must-gather.sample"
	tests := map[string]func(string, string) (string, error){
		"testdata/must-gather.zip":    ExtractZip,
		"testdata/must-gather.tar":    ExtractTar,
		"testdata/must-gather.tar.gz": ExtractTarGz,
	}

	for path, f := range tests {
		mgRootDir, err := f(path, "testdata")
		if err != nil {
			t.Error(err)
		}
		defer clearTestFiles(t)

		if want != mgRootDir {
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
