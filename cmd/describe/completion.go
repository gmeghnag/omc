package describe

import (
	"strings"

	"github.com/gmeghnag/omc/cmd/completion"
	"github.com/gmeghnag/omc/cmd/get"
	"github.com/spf13/cobra"
)

// DescribeResourceCompletionFunc provides completion for resource types and resource names
func DescribeResourceCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// If we already have a resource type argument, complete resource names
	if len(args) > 0 {
		resourceType := args[0]
		return completion.GetResourceNames(resourceType, toComplete, get.KindGroupNamespaced), cobra.ShellCompDirectiveNoFileComp
	}

	// Common resources that are describable
	commonResources := []string{
		"pod", "pods", "po",
		"node", "nodes", "no",
		"service", "services", "svc",
		"deployment", "deployments", "deploy",
		"replicaset", "replicasets", "rs",
		"statefulset", "statefulsets", "sts",
		"daemonset", "daemonsets", "ds",
		"job", "jobs",
		"cronjob", "cronjobs", "cj",
		"namespace", "namespaces", "ns",
		"persistentvolume", "persistentvolumes", "pv",
		"persistentvolumeclaim", "persistentvolumeclaims", "pvc",
		"configmap", "configmaps", "cm",
		"secret", "secrets",
		"serviceaccount", "serviceaccounts", "sa",
		"ingress", "ingresses", "ing",
		"networkpolicy", "networkpolicies", "netpol",
		"storageclass", "storageclasses", "sc",
		"clusterrole", "clusterroles",
		"clusterrolebinding", "clusterrolebindings",
		"role", "roles",
		"rolebinding", "rolebindings",
		"deploymentconfig", "deploymentconfigs", "dc",
		"buildconfig", "buildconfigs", "bc",
		"build", "builds",
		"route", "routes",
		"project", "projects",
		"imagestream", "imagestreams",
		"horizontalpodautoscaler", "horizontalpodautoscalers", "hpa",
		"endpoints", "ep",
		"event", "events", "ev",
		"replicationcontroller", "replicationcontrollers", "rc",
	}

	for _, resource := range commonResources {
		if strings.HasPrefix(resource, strings.ToLower(toComplete)) {
			completions = append(completions, resource)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
