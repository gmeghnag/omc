// Package mustgather provides discovery and merging primitives shared by the
// commands that read across several must-gathers at once.
//
// The model is: a context can point at more than one must-gather root. Roots
// are always ordered most-recent-first (by the collection timestamp written by
// `oc adm must-gather`). That ordering is the single source of truth:
//
//   - get fans out across every root, unions the results, and deduplicates by
//     metadata.uid keeping the first occurrence - so the most recent capture of
//     a resource wins.
//   - describe/logs resolve the one root a resource lives in with ResolveRoot,
//     which returns the most recent root that actually contains it, so they
//     always show the same capture get would show (no cross-capture mixing).
package mustgather

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// Root is a single must-gather root directory (the directory that directly
// contains namespaces/ and/or cluster-scoped-resources/) together with the
// collection time used to rank it against the others.
type Root struct {
	Path      string
	Timestamp time.Time
}

// timestampLayout matches the lines written in a must-gather's `timestamp`
// file, e.g. "2023-08-01 12:34:56.123456789 +0000 UTC m=+0.123". The trailing
// monotonic-clock reading (" m=...") is stripped before parsing.
const timestampLayout = "2006-01-02 15:04:05.999999999 -0700 MST"

// isRootDir reports whether dir directly holds a must-gather's resource tree.
func isRootDir(dir string) bool {
	for _, marker := range []string{"namespaces", "cluster-scoped-resources"} {
		if info, err := os.Stat(filepath.Join(dir, marker)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// DiscoverRoots walks base and returns every must-gather root found beneath it
// (base itself included), ordered most-recent-first. A single must-gather
// yields exactly one root, so callers get identical behaviour to the previous
// single-root world when base contains only one capture.
//
// Descent stops as soon as a root is found on a branch, so the per-namespace
// and per-resource subdirectories inside a must-gather are never mistaken for
// separate roots.
func DiscoverRoots(base string) ([]Root, error) {
	var roots []Root
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// A single unreadable subtree should not abort discovery of the
			// rest; skip it and continue.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if isRootDir(path) {
			roots = append(roots, Root{Path: path, Timestamp: rootTimestamp(path)})
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	SortRoots(roots)
	return roots, nil
}

// SortRoots orders roots most-recent-first. Ties (equal or missing timestamps)
// fall back to the path so the order is deterministic.
func SortRoots(roots []Root) {
	sort.SliceStable(roots, func(i, j int) bool {
		if !roots[i].Timestamp.Equal(roots[j].Timestamp) {
			return roots[i].Timestamp.After(roots[j].Timestamp)
		}
		return roots[i].Path < roots[j].Path
	})
}

// Paths returns the root paths in their current order.
func Paths(roots []Root) []string {
	paths := make([]string, len(roots))
	for i, r := range roots {
		paths[i] = r.Path
	}
	return paths
}

// rootTimestamp returns the collection time for a root. It reads the first
// usable line of the must-gather `timestamp` file (written next to the root, or
// inside it as a fallback) and, failing that, uses the root directory's
// modification time so ranking still works on archives without a timestamp.
func rootTimestamp(root string) time.Time {
	for _, p := range []string{
		filepath.Join(root, "..", "timestamp"),
		filepath.Join(root, "timestamp"),
	} {
		if t, ok := parseTimestampFile(p); ok {
			return t
		}
	}
	if info, err := os.Stat(root); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

// parseTimestampFile parses the first line of a must-gather timestamp file.
func parseTimestampFile(path string) (time.Time, bool) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.Index(line, " m="); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if t, err := time.Parse(timestampLayout, line); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// DedupByUID removes duplicate resources that share a metadata.uid, keeping the
// first occurrence of each uid. Callers pass items already ordered
// most-recent-first, so the surviving copy is the one from the most recent
// must-gather.
//
// Resources without a uid are never collapsed together: every uid-less item is
// kept. A Kubernetes object always carries a uid, but a malformed or partial
// file might not, and dropping all-but-one of them (as a naive map keyed on the
// empty string would) would silently hide real resources.
func DedupByUID(items []unstructured.Unstructured) []unstructured.Unstructured {
	if len(items) == 0 {
		return items
	}
	seen := make(map[k8stypes.UID]struct{}, len(items))
	out := make([]unstructured.Unstructured, 0, len(items))
	for _, item := range items {
		uid := item.GetUID()
		if uid == "" {
			out = append(out, item)
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, item)
	}
	return out
}

// ResolveRoot returns the first root (roots are most-recent-first) under which
// at least one of relPaths exists, together with true. It is how describe and
// logs find the single capture a resource lives in, so they read the same
// must-gather that get's uid-dedup would have selected.
//
// When nothing matches it returns the most recent root and false, letting the
// caller fall back to its usual "not found" handling against a valid root.
func ResolveRoot(roots []string, relPaths ...string) (string, bool) {
	return ResolveRootBy(roots, func(root string) bool {
		for _, rel := range relPaths {
			if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
				return true
			}
		}
		return false
	})
}

// ResolveRootBy returns the first root (most-recent-first) for which contains
// reports true, together with true. It lets a caller decide "does this capture
// hold the resource" with layout-specific logic (e.g. a pod that may live in a
// per-pod directory or inside an aggregated list) while keeping the
// most-recent-wins selection here. When none matches it returns the most recent
// root and false.
func ResolveRootBy(roots []string, contains func(root string) bool) (string, bool) {
	for _, root := range roots {
		if contains(root) {
			return root, true
		}
	}
	if len(roots) > 0 {
		return roots[0], false
	}
	return "", false
}
