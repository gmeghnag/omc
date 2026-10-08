package use

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gmeghnag/omc/types"
)

func TestIsCompressedFile(t *testing.T) {
	cases := map[string]string{
		"testdata/must-gather.tar":    fileTypeTar,
		"testdata/must-gather.tar.gz": fileTypeTarGzip,
		"testdata/must-gather.tar.xz": fileTypeXZ,
		"testdata/must-gather.zip":    fileTypeZip,
	}
	for path, wantType := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			ok, gotType, err := IsCompressedFile(path)
			if err != nil {
				t.Fatalf("IsCompressedFile(%s): %v", path, err)
			}
			if !ok || gotType != wantType {
				t.Fatalf("IsCompressedFile(%s) = (%v, %q), want (true, %q)", path, ok, gotType, wantType)
			}
		})
	}

	t.Run("plain file is not compressed", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "plain.txt")
		if err := os.WriteFile(p, []byte("just some text, definitely not an archive\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ok, _, _ := IsCompressedFile(p)
		if ok {
			t.Fatalf("expected plain file not to be detected as compressed")
		}
	})
}

func TestFindExistingContextByRootDir(t *testing.T) {
	ctxs := []types.Context{
		{Path: "/home/u/cases/123/must-gather.local.abc"},
		{Path: "/home/u/old-must-gather-backup/data"},
	}
	// Exact component match succeeds.
	if p, ok := findExistingContextByRootDir("must-gather.local.abc", ctxs); !ok || p != ctxs[0].Path {
		t.Errorf("expected match on exact root-dir component, got (%q,%v)", p, ok)
	}
	// A substring of a component must not match (previously HasSuffix/Contains
	// would have matched "must-gather" against "old-must-gather-backup").
	if p, ok := findExistingContextByRootDir("must-gather", ctxs); ok {
		t.Errorf("substring must not match a path component, got %q", p)
	}
	// Intermediate component matches too.
	if _, ok := findExistingContextByRootDir("123", ctxs); !ok {
		t.Errorf("expected match on an intermediate path component")
	}
}
