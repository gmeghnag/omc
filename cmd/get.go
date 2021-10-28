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
	"omc/cmd/helpers"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get related subcommand",
	Run: func(cmd *cobra.Command, args []string) {
		if currentContextPath == "" {
			fmt.Println("There are no must-gather resources defined.")
			os.Exit(1)
		}
		exist, _ := helpers.Exists(currentContextPath + "/namespaces")
		if !exist {
			files, err := ioutil.ReadDir(currentContextPath)
			if err != nil {
				log.Fatal(err)
			}
			quayDir := ""
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "quay") {
					quayDir = f.Name()
					currentContextPath = currentContextPath + "/" + quayDir
					break
				}
			}
			if quayDir == "" {
				fmt.Println("Some error occurred, wrong must-gather file composition")
				os.Exit(1)
			}
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
		//BUILDS
		if strings.HasPrefix(typedResource, "build") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "build" || s[0] == "build.build.openshift.io" || s[0] == "builds") {
				getBuilds(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "build" || typedResource == "build.build.openshift.io" || typedResource == "builds") {
					getBuilds(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "build" || typedResource == "build.build.openshift.io" || typedResource == "builds") {
						getBuilds(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//BUILDCONFIGS
		if strings.HasPrefix(typedResource, "bc") || strings.HasPrefix(typedResource, "buildconfig") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "bc" || s[0] == "buildconfig" || s[0] == "buildconfig.build.openshift.io" || s[0] == "buildconfigs") {
				getBuildConfigs(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "bc" || typedResource == "buildconfig" || typedResource == "buildconfig.build.openshift.io" || typedResource == "buildconfigs") {
					getBuildConfigs(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "bc" || typedResource == "buildconfig" || typedResource == "buildconfig.build.openshift.io" || typedResource == "buildconfigs") {
						getBuildConfigs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
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
		//MC
		if strings.HasPrefix(typedResource, "mc") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "mc") {
				getMachineConfig(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "mc") {
					getMachineConfig(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "mc") {
						getMachineConfig(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//MCP
		if strings.HasPrefix(typedResource, "mcp") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "mcp") {
				getMachineConfigPool(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "mcp") {
					getMachineConfigPool(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "mcp") {
						getMachineConfigPool(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
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
		//CONFIGMAP
		if strings.HasPrefix(typedResource, "cm") || strings.HasPrefix(typedResource, "configmap") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "cm" || s[0] == "configmap" || s[0] == "configmaps") {
				getConfigMaps(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "cm" || typedResource == "configmap" || typedResource == "configmaps") {
					getConfigMaps(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "cm" || typedResource == "configmap" || typedResource == "configmaps") {
						getConfigMaps(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
		//DEPLOYMENTCONFIGS
		if strings.HasPrefix(typedResource, "dc") || strings.HasPrefix(typedResource, "deploymentconfig") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "dc" || s[0] == "deploymentconfig" || s[0] == "deploymentconfig.apps.openshift.io" || s[0] == "deploymentconfigs") {
				getDeploymentConfigs(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "dc" || typedResource == "deploymentconfig" || typedResource == "deploymentconfig.apps.openshift.io" || typedResource == "deploymentconfigs") {
					getDeploymentConfigs(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "dc" || typedResource == "deploymentconfig" || typedResource == "deploymentconfig.apps.openshift.io" || typedResource == "deploymentconfigs") {
						getDeploymentConfigs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//IMAGESTREAMS
		if strings.HasPrefix(typedResource, "is") || strings.HasPrefix(typedResource, "imagestream") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "dc" || s[0] == "imagestream" || s[0] == "imagestream.imagestream.openshift.io" || s[0] == "imagestreams") {
				getImageStreams(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "is" || typedResource == "imagestream" || typedResource == "imagestream.imagestream.openshift.io" || typedResource == "imagestreams") {
					getImageStreams(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "is" || typedResource == "imagestream" || typedResource == "imagestream.imagestream.openshift.io" || typedResource == "imagestreams") {
						getImageStreams(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//JOBS
		if strings.HasPrefix(typedResource, "job") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "job" || s[0] == "jobs" || s[0] == "job.batch") {
				getJobs(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "job" || typedResource == "jobs" || typedResource == "job.batch") {
					getJobs(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "job" || typedResource == "jobs" || typedResource == "job.batch") {
						getJobs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//MACHINE
		if strings.HasPrefix(typedResource, "machine") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "machine" || s[0] == "machines" || s[0] == "machine.machine.openshift.io") {
				getMachines(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "machine" || typedResource == "machines" || typedResource == "machine.machine.openshift.io") {
					getMachines(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "machine" || typedResource == "machines" || typedResource == "machine.machine.openshift.io") {
						getMachines(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
					}

				}
			}
		}
		//MACHINESETS
		if strings.HasPrefix(typedResource, "machineset") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "machineset" || s[0] == "machinesets" || s[0] == "machineset.machine.openshift.io") {
				getMachineSets(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "machineset" || typedResource == "machinesets" || typedResource == "machineset.machine.openshift.io") {
					getMachineSets(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "machineset" || typedResource == "machinesets" || typedResource == "machineset.machine.openshift.io") {
						getMachineSets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
		//REPLICATIONCONTROLLERS
		if strings.HasPrefix(typedResource, "rc") || strings.HasPrefix(typedResource, "replicationcontroller") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "rc" || s[0] == "replicationcontroller" || s[0] == "replicationcontrollers") {
				getReplicationControllers(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "rc" || typedResource == "replicationcontroller" || typedResource == "replicationcontrollers") {
					getReplicationControllers(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "rc" || typedResource == "replicationcontroller" || typedResource == "replicationcontrollers") {
						getReplicationControllers(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
		//PROJECTS
		if strings.HasPrefix(typedResource, "project") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "project" || s[0] == "projects") {
				getProjects(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
			} else {
				if len(args) == 2 && (typedResource == "project" || typedResource == "projects") {
					getProjects(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
				} else {
					if len(args) == 1 && (typedResource == "project" || typedResource == "projects") {
						getProjects(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate)
					}

				}
			}
		}
		//SECRETS
		if strings.HasPrefix(typedResource, "secret") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "secret" || s[0] == "secrets") {
				getSecrets(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "secret" || typedResource == "secrets") {
					getSecrets(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "secret" || typedResource == "secrets") {
						getSecrets(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
		//PVC
		if strings.HasPrefix(typedResource, "pvc") || strings.HasPrefix(typedResource, "persistentvolumeclaim") {
			if s := strings.Split(typedResource, "/"); len(s) == 2 && (s[0] == "pvc" || s[0] == "persistentvolumeclaim" || s[0] == "persistentvolumeclaims") {
				getPersistentVolumeClaims(currentContextPath, defaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			} else {
				if len(args) == 2 && (typedResource == "pvc" || typedResource == "persistentvolumeclaim" || typedResource == "persistentvolumeclaims") {
					getPersistentVolumeClaims(currentContextPath, defaultConfigNamespace, args[1], allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
				} else {
					if len(args) == 1 && (typedResource == "pvc" || typedResource == "persistentvolumeclaim" || typedResource == "persistentvolumeclaims") {
						getPersistentVolumeClaims(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
			empty = getReplicationControllers(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
			empty = getJobs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getDeploymentConfigs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getBuildConfigs(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getBuilds(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
			if !empty {
				fmt.Println("")
			}
			empty = getImageStreams(currentContextPath, defaultConfigNamespace, "", allNamespacesFlag, outputFlag, showLabels, jsonPathTemplate, allResources)
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
