package logs

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// LogsCompletionFunc provides completion for pod names
func LogsCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// logs command takes pod name as first argument
	// If we already have a pod name, no more completions
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Check if we have a must-gather path configured
	if vars.MustGatherRootPath == "" {
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Get namespace from flag or use default
	namespace := vars.Namespace
	if namespace == "" {
		namespace = vars.DefaultProject
	}

	// Get pod names from must-gather
	podNames := getPodNames(namespace, toComplete)
	return podNames, cobra.ShellCompDirectiveNoFileComp
}

// getPodNames returns pod names from the must-gather
func getPodNames(namespace, toComplete string) []string {
	var podNames []string

	if namespace != "" {
		// Get pods from specific namespace
		podNames = getPodsFromNamespace(namespace, toComplete)
	} else {
		// Get pods from all namespaces
		podNames = getPodsFromAllNamespaces(toComplete)
	}

	return podNames
}

// getPodsFromNamespace gets pod names from a specific namespace
func getPodsFromNamespace(namespace, toComplete string) []string {
	var podNames []string
	
	// Check for aggregated pods.yaml file
	podsFile := filepath.Join(vars.MustGatherRootPath, "namespaces", namespace, "core", "pods.yaml")
	names := extractPodNamesFromFile(podsFile, toComplete)
	podNames = append(podNames, names...)

	// Check for individual pod files in pods/ directory
	podsDir := filepath.Join(vars.MustGatherRootPath, "namespaces", namespace, "core", "pods")
	names = extractPodNamesFromDir(podsDir, toComplete)
	podNames = append(podNames, names...)

	return podNames
}

// getPodsFromAllNamespaces gets pod names from all namespaces
func getPodsFromAllNamespaces(toComplete string) []string {
	var podNames []string

	namespacesPath := filepath.Join(vars.MustGatherRootPath, "namespaces")
	namespaces, err := ioutil.ReadDir(namespacesPath)
	if err != nil {
		return podNames
	}

	for _, ns := range namespaces {
		if ns.IsDir() {
			names := getPodsFromNamespace(ns.Name(), toComplete)
			podNames = append(podNames, names...)
		}
	}

	return podNames
}

// extractPodNamesFromFile extracts pod names from a YAML file containing a list
func extractPodNamesFromFile(filePath, toComplete string) []string {
	var podNames []string

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return podNames
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name string `yaml:"name"`
			} `yaml:"metadata"`
		} `yaml:"items"`
	}

	err = yaml.Unmarshal(data, &podList)
	if err != nil {
		return podNames
	}

	for _, pod := range podList.Items {
		podName := pod.Metadata.Name
		if strings.HasPrefix(podName, toComplete) {
			podNames = append(podNames, podName)
		}
	}

	return podNames
}

// extractPodNamesFromDir extracts pod names from individual YAML files in a directory
func extractPodNamesFromDir(dirPath, toComplete string) []string {
	var podNames []string

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return podNames
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		podName := strings.TrimSuffix(file.Name(), ".yaml")
		if strings.HasPrefix(podName, toComplete) {
			podNames = append(podNames, podName)
		}
	}

	return podNames
}
