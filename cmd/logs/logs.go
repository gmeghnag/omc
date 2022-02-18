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
package logs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var LogLevel string

// logsCmd represents the logs command
var Logs = &cobra.Command{
	Use:   "logs",
	Short: "Print the logs for a container in a pod",
	Run: func(cmd *cobra.Command, args []string) {
		if vars.MustGatherRootPath == "" {
			fmt.Println("There are no must-gather resources defined.")
			os.Exit(1)
		}
		exist, _ := helpers.Exists(vars.MustGatherRootPath + "/namespaces")
		if !exist {
			files, err := ioutil.ReadDir(vars.MustGatherRootPath)
			if err != nil {
				log.Fatal(err)
			}
			var QuayString string
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "quay") {
					QuayString = f.Name()
					vars.MustGatherRootPath = vars.MustGatherRootPath + "/" + QuayString
					break
				}
			}
			if QuayString == "" {
				fmt.Println("Some error occurred, wrong must-gather file composition")
				os.Exit(1)
			}
		}
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" {
			vars.Namespace = namespaceFlag
		}
		podName := ""
		containerName, _ := cmd.Flags().GetString("container")
		previousFlag, _ := cmd.Flags().GetBool("previous")
		allContainersFlag, _ := cmd.Flags().GetBool("all-containers")
		logLevels := []string{}
		if LogLevel != "" {
			logLevels = strings.Split(LogLevel, ",")
		}

		if len(args) == 0 || len(args) > 2 {
			fmt.Println("error: expected 'logs [-p] (POD | TYPE/NAME) [-c CONTAINER]'.")
			fmt.Println("POD or TYPE/NAME is a required argument for the logs command")
			fmt.Println("See 'omc logs -h' for help and examples")
			os.Exit(1)
		}
		if len(args) == 1 {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				podName = s[1]
				if podName == "" {
					fmt.Println("error: arguments in resource/name form must have a single resource and name")
					os.Exit(1)
				}
				logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, allContainersFlag, logLevels)
			} else {
				podName = s[0]
				logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, allContainersFlag, logLevels)
			}
		}
		if len(args) == 2 {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				if containerName != "" {
					fmt.Println("error: only one of -c or an inline [CONTAINER] arg is allowed")
					os.Exit(1)
				} else {
					podName = s[1]
					if podName == "" {
						fmt.Println("error: arguments in resource/name form must have a single resource and name")
						os.Exit(1)
					}
					containerName = args[1]
					logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, allContainersFlag, logLevels)
				}
			} else {
				if containerName != "" {
					fmt.Println("error: only one of -c or an inline [CONTAINER] arg is allowed")
					os.Exit(1)
				} else {
					podName = args[0]
					containerName = args[1]
					logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, allContainersFlag, logLevels)
				}
			}
		}
	},
}

func init() {
	Logs.PersistentFlags().StringVarP(&vars.Container, "container", "c", "", "Print the logs of this container")
	Logs.PersistentFlags().BoolVarP(&vars.Previous, "previous", "p", false, "Print the logs for the previous instance of the container in a pod if it exists.")
	Logs.PersistentFlags().BoolVarP(&vars.AllContainers, "all-containers", "", false, "Get all containers' logs in the pod(s).")
	Logs.Flags().StringVarP(&LogLevel, "log-level", "l", "", "Filter logs by level (info|error|worning), you can filter for more concatenating them comma separated.")
}
