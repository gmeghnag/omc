package stern

import (
	"fmt"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// normalizeMustGatherRoot ensures vars.MustGatherRootPath points at the
// directory that directly contains "namespaces". Some must-gathers (e.g.
// quay-prefixed ODF captures) nest the expected layout one directory down.
func normalizeMustGatherRoot() error {
	if vars.MustGatherRootPath == "" {
		return fmt.Errorf("there are no must-gather resources defined")
	}
	if _, err := os.Stat(vars.MustGatherRootPath + "/namespaces"); err == nil {
		return nil
	}
	entries, err := os.ReadDir(vars.MustGatherRootPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "quay") {
			vars.MustGatherRootPath = vars.MustGatherRootPath + "/" + entry.Name()
			return nil
		}
	}
	return fmt.Errorf("wrong must-gather file composition")
}

// listPods returns all pods found under a namespace directory within a
// must-gather. helpers.PodObjects resolves the aggregated core/pods.yaml and
// both per-pod layouts (pods/<name>/<name>.yaml and core/pods/<name>.yaml),
// which must-gather images use interchangeably.
func listPods(namespacePath string) ([]corev1.Pod, error) {
	objects, err := helpers.PodObjects(namespacePath)
	if err != nil {
		return nil, err
	}
	var pods []corev1.Pod
	for _, object := range objects {
		var pod corev1.Pod
		if err := yaml.Unmarshal(object.Data, &pod); err != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", object.Path, err)
		}
		pods = append(pods, pod)
	}
	return pods, nil
}
