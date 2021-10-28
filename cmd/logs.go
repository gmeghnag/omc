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
	"strings"

	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Print the logs for a container in a pod",
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
		}
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" {
			defaultConfigNamespace = namespaceFlag
		}
		podName := ""
		containerName, _ := cmd.Flags().GetString("container")
		previousFlag, _ := cmd.Flags().GetBool("previous")
		allContainersFlag, _ := cmd.Flags().GetBool("all-containers")

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
				logsPods(currentContextPath, defaultConfigNamespace, podName, containerName, previousFlag, allContainersFlag)
			} else {
				podName = s[0]
				logsPods(currentContextPath, defaultConfigNamespace, podName, containerName, previousFlag, allContainersFlag)
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
					logsPods(currentContextPath, defaultConfigNamespace, podName, containerName, previousFlag, allContainersFlag)
				}
			} else {
				if containerName != "" {
					fmt.Println("error: only one of -c or an inline [CONTAINER] arg is allowed")
					os.Exit(1)
				} else {
					podName = args[0]
					containerName = args[1]
					logsPods(currentContextPath, defaultConfigNamespace, podName, containerName, previousFlag, allContainersFlag)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.PersistentFlags().StringVarP(&output, "container", "c", "", "Print the logs of this container")
	logsCmd.PersistentFlags().BoolP("previous", "p", false, "Print the logs for the previous instance of the container in a pod if it exists.")
	logsCmd.PersistentFlags().BoolP("all-containers", "", false, "Get all containers' logs in the pod(s).")
}
