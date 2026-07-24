package stern

import (
	"fmt"
	"os"
	"strings"

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
// must-gather. It prefers the namespace's core/pods.yaml, falling back to
// reading each pod's individual manifest under
// <namespacePath>/pods/<name>/<name>.yaml when that file is missing or empty.
func listPods(namespacePath string) ([]corev1.Pod, error) {
	podsYamlPath := namespacePath + "/core/pods.yaml"
	if data, err := os.ReadFile(podsYamlPath); err == nil {
		var podList corev1.PodList
		if err := yaml.Unmarshal(data, &podList); err != nil {
			return nil, fmt.Errorf("error unmarshaling %s: %w", podsYamlPath, err)
		}
		if len(podList.Items) > 0 {
			return podList.Items, nil
		}
	}

	podsDir := namespacePath + "/pods"
	entries, err := os.ReadDir(podsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", podsDir, err)
	}
	var pods []corev1.Pod
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		podYamlPath := podsDir + "/" + entry.Name() + "/" + entry.Name() + ".yaml"
		data, err := os.ReadFile(podYamlPath)
		if err != nil {
			continue
		}
		var pod corev1.Pod
		if err := yaml.Unmarshal(data, &pod); err != nil {
			continue
		}
		pods = append(pods, pod)
	}
	return pods, nil
}
