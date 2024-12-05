package get

import (
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
)

func TestReadDirForResources(t *testing.T) {
	tests := []struct {
		name     string
		in       fstest.MapFS
		expected []string
	}{
		{
			name: "read correct resource files/dirs",
			in: fstest.MapFS{
				"resource-file-1.yaml":            {Data: []byte("abc")},
				"resource.yaml":                   {Data: []byte("abc")},
				"1.yaml":                          {Data: []byte("abc")},
				"resource.with.dot.filename.yaml": {Data: []byte("abc")},
				"resource-directory-name":         {Data: []byte("abc"), Mode: fs.ModeDir},
				"resource.directory.with.dot":     {Data: []byte("abc"), Mode: fs.ModeDir},
			},
			expected: []string{
				"resource-file-1.yaml",
				"resource.yaml",
				"1.yaml",
				"resource.with.dot.filename.yaml",
				"resource-directory-name",
				"resource.directory.with.dot",
			},
		},
		{
			name: "read only resource files/dirs matching the expected name convention",
			in: fstest.MapFS{
				"resource-file-1.yaml":             {Data: []byte("abc")},
				"._faulthy-resource-filename.yaml": {Data: []byte("abc")}, // e.g. AppleDouble encoded Macintosh file
				".resource-filename.yaml.swp":      {Data: []byte("abc")},
				"-resource-filename.yaml":          {Data: []byte("abc")},
			},
			expected: []string{"resource-file-1.yaml"},
		},
		{
			name: "read only resource files/dir with size > 0",
			in: fstest.MapFS{
				"resource-file-1.yaml":         {Data: []byte("abc")},
				"empty-resource-filename.yaml": {},
			},
			expected: []string{"resource-file-1.yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := readDirForResources(tc.in)
			if len(got) != len(tc.expected) {
				t.Errorf("Got: %v \n", got)
				t.Errorf("Want: %v \n", tc.expected)
			} else {
				gotNames := make([]string, 0)
				for _, dir := range got {
					gotNames = append(gotNames, dir.Name())
				}
				sort.Slice(gotNames, func(i, j int) bool {
					return gotNames[i] > gotNames[j]
				})
				sort.Slice(tc.expected, func(i, j int) bool {
					return tc.expected[i] > tc.expected[j]
				})
				if !reflect.DeepEqual(gotNames, tc.expected) {
					t.Error("Got:")
					for _, g := range gotNames {
						t.Error(g)
					}
					t.Error("Want:")
					for _, te := range tc.expected {
						t.Error(te)
					}
				}
			}
		})
	}
}
