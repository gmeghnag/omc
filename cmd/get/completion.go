package get

import (
	"strings"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

// GetResourceCompletionFunc provides completion for resource types
func GetResourceCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// If we already have an argument, don't offer more resource types
	// (we're now completing resource names, which we can't do with must-gather)
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get all known resource types from vars.KnownResources
	seen := make(map[string]bool)
	for resourceKey := range vars.KnownResources {
		resourceLower := strings.ToLower(resourceKey)

		// Only add if it starts with what user has typed
		if strings.HasPrefix(resourceLower, strings.ToLower(toComplete)) {
			if !seen[resourceLower] {
				seen[resourceLower] = true
				completions = append(completions, resourceLower)
			}
		}
	}

	// Common short forms that might not be in KnownResources
	commonShortForms := []string{
		"po", "pod", "pods",
		"deploy", "deployment", "deployments",
		"svc", "service", "services",
		"ns", "namespace", "namespaces",
		"no", "node", "nodes",
		"pv", "persistentvolume", "persistentvolumes",
		"pvc", "persistentvolumeclaim", "persistentvolumeclaims",
		"cm", "configmap", "configmaps",
		"secret", "secrets",
		"sa", "serviceaccount", "serviceaccounts",
		"deploy", "deployment", "deployments",
		"ds", "daemonset", "daemonsets",
		"rs", "replicaset", "replicasets",
		"sts", "statefulset", "statefulsets",
		"job", "jobs",
		"cj", "cronjob", "cronjobs",
		"ing", "ingress", "ingresses",
		"netpol", "networkpolicy", "networkpolicies",
		"pdb", "poddisruptionbudget", "poddisruptionbudgets",
		"sc", "storageclass", "storageclasses",
		"crd", "crds", "customresourcedefinition", "customresourcedefinitions",
		"clusterrole", "clusterroles",
		"clusterrolebinding", "clusterrolebindings",
		"role", "roles",
		"rolebinding", "rolebindings",
		"dc", "deploymentconfig", "deploymentconfigs",
		"bc", "buildconfig", "buildconfigs",
		"build", "builds",
		"route", "routes",
		"project", "projects",
		"imagestream", "imagestreams",
		"imagestreamtag", "imagestreamtags",
		"template", "templates",
		"hpa", "horizontalpodautoscaler", "horizontalpodautoscalers",
		"ep", "endpoints",
		"ev", "event", "events",
		"rc", "replicationcontroller", "replicationcontrollers",
		"quota", "resourcequota", "resourcequotas",
		"limits", "limitrange", "limitranges",
		"pc", "priorityclass", "priorityclasses",
		"csr", "certificatesigningrequest", "certificatesigningrequests",
		"lease", "leases",
		"endpointslice", "endpointslices",
		"apiservice", "apiservices",
		"scc", "securitycontextconstraints",
	}

	for _, resource := range commonShortForms {
		if strings.HasPrefix(resource, strings.ToLower(toComplete)) {
			if !seen[resource] {
				seen[resource] = true
				completions = append(completions, resource)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
