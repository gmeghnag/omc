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
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var validArgs []string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get related subcommand",
	Run: func(cmd *cobra.Command, args []string) {
		if currentContextPath == "" {
			fmt.Println("There are no must-gather resources defined.")
			os.Exit(1)
		}
		files, err := ioutil.ReadDir(currentContextPath)
		if err != nil {
			log.Fatal(err)
		}
		var QuayString string
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "quay") {
				QuayString = f.Name()
				currentContextPath = currentContextPath + "/" + QuayString
				break
			}
		}
		if QuayString == "" {
			fmt.Println("Some error occurred, wrong must-gather file composition")
			os.Exit(1)
		}
		allNamespacesFlag, _ := cmd.Flags().GetBool("all-namespaces")
		showLabels, _ := cmd.Flags().GetBool("show-labels")
		outputFlag, _ := cmd.Flags().GetString("output")
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" && !allNamespacesFlag {
			defaultConfigNamespace, _ = rootCmd.PersistentFlags().GetString("namespace")
		}

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
		typedResource := strings.ToLower(args[0])
		//CLUSTERVERSION
		if strings.HasPrefix(typedResource, "clusterversion") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "clusterversion" || s[0] == "clusterversions") {
				getClusterVersion(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "clusterversion" || typedResource == "clusterversions") {
					getClusterVersion(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "clusterversion" || typedResource == "clusterversions") {
						getClusterVersion(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//CLUSTEROPERATORS
		if strings.HasPrefix(typedResource, "co") || strings.HasPrefix(typedResource, "clusteroperator") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "co" || s[0] == "clusteroperator" || s[0] == "clusteroperators") {
				getClusterOperators(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "co" || typedResource == "clusteroperator" || typedResource == "clusteroperators") {
					getClusterOperators(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "co" || typedResource == "clusteroperator" || typedResource == "clusteroperators") {
						getClusterOperators(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//EVENTS
		if strings.HasPrefix(typedResource, "event") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "event" || s[0] == "events") {
				getEvents(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "event" || typedResource == "events") {
					getEvents(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "event" || typedResource == "events") {
						getEvents(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//DAEMONSET
		if strings.HasPrefix(typedResource, "daemonset") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "daemonset" || s[0] == "daemonset.apps" || s[0] == "daemonsts") {
				getDaemonsets(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "daemonset" || typedResource == "daemonset.apps" || typedResource == "daemonsets") {
					getDaemonsets(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "daemonset" || typedResource == "daemonset.apps" || typedResource == "daemonsets") {
						getDaemonsets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//DEPLOYMENTS
		if strings.HasPrefix(typedResource, "deployment") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "deployment" || s[0] == "deployment.apps" || s[0] == "deployments") {
				getDeployments(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "deployment" || typedResource == "deployment.apps" || typedResource == "deployments") {
					getDeployments(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "deployment" || typedResource == "deployment.apps" || typedResource == "deployments") {
						getDeployments(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//REPLICASETS
		if strings.HasPrefix(typedResource, "rs") || strings.HasPrefix(typedResource, "replicaset") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "rs" || s[0] == "replicaset" || s[0] == "replicaset.apps" || s[0] == "replicasets") {
				getReplicaSets(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "rs" || typedResource == "replicaset" || typedResource == "replicaset.apps" || typedResource == "replicasets") {
					getReplicaSets(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "rs" || typedResource == "replicaset" || typedResource == "replicaset.apps" || typedResource == "replicasets") {
						getReplicaSets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//PODS
		if strings.HasPrefix(typedResource, "po") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				getPods(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "po" || typedResource == "pod" || typedResource == "pods") {
					getPods(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "po" || typedResource == "pod" || typedResource == "pods") {
						getPods(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//SERVICES
		if strings.HasPrefix(typedResource, "svc") || strings.HasPrefix(typedResource, "service") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "svc" || s[0] == "service" || s[0] == "services") {
				getServices(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "svc" || typedResource == "service" || typedResource == "services") {
					getServices(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "svc" || typedResource == "service" || typedResource == "services") {
						getServices(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//ROUTE
		if strings.HasPrefix(typedResource, "route") || strings.HasPrefix(typedResource, "routes") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "route.route.openshift.io" || s[0] == "route" || s[0] == "routes") {
				getRoutes(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "route.route.openshift.io" || typedResource == "route" || typedResource == "routes") {
					getRoutes(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "route.route.openshift.io" || typedResource == "route" || typedResource == "routes") {
						getRoutes(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//NODES
		if strings.HasPrefix(typedResource, "node") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "node" || s[0] == "nodes") {
				getNodes(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "node" || typedResource == "nodes") {
					getNodes(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "node" || typedResource == "nodes") {
						getNodes(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//PV
		if strings.HasPrefix(typedResource, "pv") || strings.HasPrefix(typedResource, "persistentvolume") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "pv" || s[0] == "persistentvolume" || s[0] == "persistentvolumes") {
				getPersistentVolumes(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "pv" || typedResource == "persistentvolume" || typedResource == "persistentvolumes") {
					getPersistentVolumes(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "pv" || typedResource == "persistentvolume" || typedResource == "persistentvolumes") {
						getPersistentVolumes(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//SC
		if strings.HasPrefix(typedResource, "sc") || strings.HasPrefix(typedResource, "storageclass") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "sc" || s[0] == "storageclass" || s[0] == "storageclasses") {
				getStorageClasses(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "sc" || typedResource == "storageclass" || typedResource == "storageclasses") {
					getStorageClasses(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "sc" || typedResource == "storageclass" || typedResource == "storageclasses") {
						getStorageClasses(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		if len(args) == 1 && typedResource == "all" {
			allResources = true
			empty := getPods(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getServices(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getDaemonsets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getDeployments(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getReplicaSets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getRoutes(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().BoolP("all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	getCmd.PersistentFlags().BoolP("show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	getCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	//getCmd.PersistentFlags().StringVarP(&selector, "selector", "l", "", "elector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
}
