/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		allNamespacesFlag, _ := cmd.Flags().GetBool("all-namespaces")
		showLabels, _ := cmd.Flags().GetBool("show-labels")
		outputFlag, _ := cmd.Flags().GetString("output")
		//namespace, _ := rootCmd.PersistentFlags().GetString("namespace")
		allResources := false
		jsonPathTemplate := ""
		if strings.HasPrefix(outputFlag, "jsonpath=") {
			s := outputFlag[9:]
			if len(s) < 1 {
				fmt.Println("error: template format specified but no template given")
				os.Exit(1)
			}
			jsonPathTemplate = s
		}
		if len(args) == 0 || len(args) > 2 {
			fmt.Println("Expected one or two arguments, found: " + strconv.Itoa(len(args)) + ".")
			os.Exit(1)
		}
		//CLUSTEROPERATORS
		if strings.HasPrefix(args[0], "co") || strings.HasPrefix(args[0], "clusteroperator") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "co" || s[0] == "clusteroperator") {
				getClusterOperators(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (args[0] == "co" || args[0] == "clusteroperator") {
					getClusterOperators(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (args[0] == "co" || args[0] == "clusteroperator") {
						getClusterOperators(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//EVENTS
		if strings.HasPrefix(args[0], "event") || strings.HasPrefix(args[0], "events") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "event" || s[0] == "events") {
				getEvents(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (args[0] == "event" || args[0] == "events") {
					getEvents(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (args[0] == "event" || args[0] == "events") {
						getEvents(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//DEPLOYMENTS
		if strings.HasPrefix(args[0], "deployment") || strings.HasPrefix(args[0], "deployments") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "deployment" || s[0] == "deployments") {
				getDeployments(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (args[0] == "deployment" || args[0] == "deployments") {
					getDeployments(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (args[0] == "deployment" || args[0] == "deployments") {
						getDeployments(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//REPLICASETS
		if strings.HasPrefix(args[0], "rs") || strings.HasPrefix(args[0], "replicaset") || strings.HasPrefix(args[0], "replicasets") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "rs" || s[0] == "replicaset" || s[0] == "replicasets") {
				getReplicaSets(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (args[0] == "rs" || args[0] == "replicaset" || args[0] == "replicasets") {
					getReplicaSets(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (args[0] == "rs" || args[0] == "replicaset" || args[0] == "replicasets") {
						getReplicaSets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//PODS
		if strings.HasPrefix(args[0], "pod") || strings.HasPrefix(args[0], "pods") || strings.HasPrefix(args[0], "po") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				getPods(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (args[0] == "po" || args[0] == "pod" || args[0] == "pods") {
					getPods(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (args[0] == "po" || args[0] == "pod" || args[0] == "pods") {
						getPods(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//SERVICES
		if strings.HasPrefix(args[0], "svc") || strings.HasPrefix(args[0], "service") || strings.HasPrefix(args[0], "services") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "svc" || s[0] == "service" || s[0] == "services") {
				getServices(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (args[0] == "svc" || args[0] == "service" || args[0] == "services") {
					getServices(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (args[0] == "svc" || args[0] == "service" || args[0] == "services") {
						getServices(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}

		//NODES
		if strings.HasPrefix(args[0], "node") || strings.HasPrefix(args[0], "nodes") {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "node" || s[0] == "nodes") {
				getNodes(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (args[0] == "node" || args[0] == "nodes") {
					getNodes(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (args[0] == "node" || args[0] == "nodes") {
						getNodes(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		if len(args) == 1 && args[0] == "all" {
			allResources = true
			empty := getPods(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getServices(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getDeployments(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
		}
		//else {
		//	fmt.Println("No resources found in " + namespace + " namespace")
		//}
	},
}

func init() {
	//fmt.Println("inside get init")

	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().BoolP("all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	getCmd.PersistentFlags().BoolP("show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	getCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
