package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// PodObject is one pod manifest read from a must-gather, with the file it came
// from so callers can report a precise path on decode errors.
type PodObject struct {
	Path string
	Data []byte
}

// PodObjects returns the manifest of every pod captured for one namespace. It
// is the single place that knows the pod storage layouts, because must-gather
// images do not agree on them:
//
//   - aggregated: <ns>/core/pods.yaml, a List of pods
//   - nested:     <ns>/pods/<pod>/<pod>.yaml
//   - flat:       <ns>/core/pods/<pod>.yaml
//
// They are tried in that order and the first layout holding at least one pod
// wins, so a populated aggregated list is never mixed with per-pod files and
// the two per-pod layouts are never merged. A namespace with none of them
// yields an empty slice and no error.
//
// Each Data value is a single pod, either YAML or JSON; decode it with
// sigs.k8s.io/yaml into corev1.Pod or unstructured.Unstructured as needed.
func PodObjects(namespacePath string) ([]PodObject, error) {
	aggregated := filepath.Join(namespacePath, "core", "pods.yaml")
	data, err := os.ReadFile(aggregated)
	switch {
	case err == nil:
		var list struct {
			Items []json.RawMessage `json:"items"`
		}
		if uerr := yaml.Unmarshal(data, &list); uerr != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", aggregated, uerr)
		}
		if len(list.Items) > 0 {
			objects := make([]PodObject, 0, len(list.Items))
			for _, item := range list.Items {
				objects = append(objects, PodObject{Path: aggregated, Data: item})
			}
			return objects, nil
		}
		// An empty (or item-less) aggregated file does not raise an unmarshal
		// error, so the count check above is what actually detects it.
	case !os.IsNotExist(err):
		return nil, fmt.Errorf("error reading %s: %w", aggregated, err)
	}

	for _, layout := range []struct {
		dir    string
		nested bool
	}{
		{filepath.Join(namespacePath, "pods"), true},
		{filepath.Join(namespacePath, "core", "pods"), false},
	} {
		objects, perr := podObjectsPerFile(layout.dir, layout.nested)
		if perr != nil {
			return nil, perr
		}
		if len(objects) > 0 {
			return objects, nil
		}
	}
	return nil, nil
}

// PodExists reports whether the namespace captured the named pod in any of the
// layouts PodObjects knows about. Unlike PodObjects, which returns the pods of
// the first populated layout, this consults all of them: callers use it to pick
// the must-gather a name resolves to, and a capture may hold, say, a populated
// aggregated list plus a per-pod file for a pod the list got wrong.
func PodExists(namespacePath, name string) bool {
	for _, podPath := range []string{
		filepath.Join(namespacePath, "pods", name, name+".yaml"),
		filepath.Join(namespacePath, "core", "pods", name+".yaml"),
	} {
		if fi, err := os.Stat(podPath); err == nil && fi.Mode().IsRegular() && fi.Size() > 0 {
			return true
		}
	}

	aggregated := filepath.Join(namespacePath, "core", "pods.yaml")
	data, err := os.ReadFile(aggregated)
	if err != nil {
		return false
	}
	var list struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := yaml.Unmarshal(data, &list); err != nil {
		klog.V(3).ErrorS(err, "Failed to unmarshal pods list", "path", aggregated)
		return false
	}
	for _, item := range list.Items {
		if item.Metadata.Name == name {
			return true
		}
	}
	return false
}

// podObjectsPerFile reads a per-pod layout: nested reads <dir>/<pod>/<pod>.yaml,
// flat reads <dir>/<pod>.yaml. Anything that is not a pod manifest is skipped -
// in particular the container log directories that must-gathers store next to
// the manifests (core/pods/<pod>/<container>/logs) - so a log-only path never
// fails the read nor renders as a pod.
func podObjectsPerFile(dir string, nested bool) ([]PodObject, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading %s: %w", dir, err)
	}
	var objects []PodObject
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		var podPath string
		if nested {
			if !entry.IsDir() {
				continue
			}
			podPath = filepath.Join(dir, name, name+".yaml")
		} else {
			if entry.IsDir() || filepath.Ext(name) != ".yaml" {
				continue
			}
			podPath = filepath.Join(dir, name)
		}
		_file, err := os.ReadFile(podPath)
		if err != nil {
			if os.IsNotExist(err) { // e.g. a container log directory
				klog.V(3).Infof("no pod manifest at %s, skipping", podPath)
				continue
			}
			return nil, fmt.Errorf("error reading %s: %w", podPath, err)
		}
		if len(_file) == 0 {
			continue
		}
		objects = append(objects, PodObject{Path: podPath, Data: _file})
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].Path < objects[j].Path })
	return objects, nil
}
