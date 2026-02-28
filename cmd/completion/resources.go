package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gmeghnag/omc/vars"
	"sigs.k8s.io/yaml"
)

// KindGroupNamespacedFunc is a function type for resolving resource metadata
type KindGroupNamespacedFunc func(string) (string, string, string, bool, error)

// GetResourceNames returns a list of resource names for the given resource type
// The kindGroupNamespacedFunc parameter should be get.KindGroupNamespaced
func GetResourceNames(resourceType, toComplete string, kindGroupNamespacedFunc KindGroupNamespacedFunc) []string {
	var completions []string

	// Check if we have a must-gather path configured
	if vars.MustGatherRootPath == "" {
		return completions
	}

	// Determine the resource plural, group, and whether it's namespaced
	resourceNamePlural, resourceGroup, _, namespaced, err := kindGroupNamespacedFunc(resourceType)
	if err != nil {
		return completions
	}

	// Get namespace from flag or use current namespace
	namespace := vars.Namespace
	if namespace == "" {
		namespace = vars.DefaultProject
	}

	var resourceNames []string

	if namespaced {
		// For namespaced resources
		if namespace != "" {
			// Get resources from specific namespace
			resourceNames = getResourceNamesFromNamespace(resourceNamePlural, resourceGroup, namespace, toComplete)
		} else {
			// Get resources from all namespaces
			resourceNames = getResourceNamesFromAllNamespaces(resourceNamePlural, resourceGroup, toComplete)
		}
	} else {
		// For cluster-scoped resources
		resourceNames = getClusterScopedResourceNames(resourceNamePlural, resourceGroup, toComplete)
	}

	return resourceNames
}

// getResourceNamesFromNamespace gets resource names from a specific namespace
func getResourceNamesFromNamespace(resourcePlural, resourceGroup, namespace, toComplete string) []string {
	var names []string

	resourcesPath := fmt.Sprintf("%s/namespaces/%s/%s/%s.yaml", vars.MustGatherRootPath, namespace, resourceGroup, resourcePlural)

	// Try reading the aggregated file first
	if fileInfo, err := os.Stat(resourcesPath); err == nil && fileInfo.Size() > 0 {
		names = extractResourceNamesFromFile(resourcesPath, toComplete)
	} else {
		// Try reading individual files from directory
		resourceDir := fmt.Sprintf("%s/namespaces/%s/%s/%s", vars.MustGatherRootPath, namespace, resourceGroup, resourcePlural)
		names = extractResourceNamesFromDir(resourceDir, toComplete)
	}

	return names
}

// getResourceNamesFromAllNamespaces gets resource names from all namespaces
func getResourceNamesFromAllNamespaces(resourcePlural, resourceGroup, toComplete string) []string {
	var names []string
	seen := make(map[string]bool)

	namespacesPath := fmt.Sprintf("%s/namespaces", vars.MustGatherRootPath)
	namespaces, err := os.ReadDir(namespacesPath)
	if err != nil {
		return names
	}

	for _, ns := range namespaces {
		if !ns.IsDir() {
			continue
		}

		nsNames := getResourceNamesFromNamespace(resourcePlural, resourceGroup, ns.Name(), toComplete)
		for _, name := range nsNames {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	return names
}

// getClusterScopedResourceNames gets cluster-scoped resource names
func getClusterScopedResourceNames(resourcePlural, resourceGroup, toComplete string) []string {
	var names []string

	resourcesPath := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s.yaml", vars.MustGatherRootPath, resourceGroup, resourcePlural)

	// Try reading the aggregated file first
	if fileInfo, err := os.Stat(resourcesPath); err == nil && fileInfo.Size() > 0 {
		names = extractResourceNamesFromFile(resourcesPath, toComplete)
	} else {
		// Try reading individual files from directory
		resourceDir := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s", vars.MustGatherRootPath, resourceGroup, resourcePlural)
		names = extractResourceNamesFromDir(resourceDir, toComplete)
	}

	return names
}

// extractResourceNamesFromFile extracts resource names from a YAML file containing a list
func extractResourceNamesFromFile(filePath, toComplete string) []string {
	var names []string

	data, err := os.ReadFile(filePath)
	if err != nil {
		return names
	}

	// Parse as a list
	var resourceList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := yaml.Unmarshal(data, &resourceList); err != nil {
		return names
	}

	for _, item := range resourceList.Items {
		if item.Metadata.Name != "" && strings.HasPrefix(item.Metadata.Name, toComplete) {
			names = append(names, item.Metadata.Name)
		}
	}

	return names
}

// extractResourceNamesFromDir extracts resource names from individual YAML files in a directory
func extractResourceNamesFromDir(dirPath, toComplete string) []string {
	var names []string

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return names
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Extract name from filename (usually <name>.yaml)
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		if strings.HasPrefix(name, toComplete) {
			names = append(names, name)
		}
	}

	return names
}
