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
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var LogLevel string

// logsCmd represents the logs command
var Logs = &cobra.Command{
	Use:          "logs",
	Short:        "Print the logs for a container in a pod",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if vars.MustGatherRootPath == "" {
			return fmt.Errorf("there are no must-gather resources defined")
		}
		exist, _ := helpers.Exists(vars.MustGatherRootPath + "/namespaces")
		if !exist {
			files, err := os.ReadDir(vars.MustGatherRootPath)
			if err != nil {
				return err
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
				return fmt.Errorf("wrong must-gather file composition")
			}
		}
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" {
			vars.Namespace = namespaceFlag
		}
		podName := ""
		containerName, _ := cmd.Flags().GetString("container")
		previousFlag, _ := cmd.Flags().GetBool("previous")
		rotatedFlag, _ := cmd.Flags().GetBool("rotated")
		insecureFlag, _ := cmd.Flags().GetBool("insecure")
		allContainersFlag, _ := cmd.Flags().GetBool("all-containers")
		logLevels := []string{}
		if LogLevel != "" {
			logLevels = strings.Split(LogLevel, ",")
		}

		if len(args) == 0 || len(args) > 2 {
			return fmt.Errorf("expected 'logs [-p] (POD | TYPE/NAME) [-c CONTAINER]'; POD or TYPE/NAME is a required argument for the logs command")
		}
		if len(args) == 1 {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				podName = s[1]
				if podName == "" {
					return fmt.Errorf("arguments in resource/name form must have a single resource and name")
				}
				return logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, rotatedFlag, allContainersFlag, logLevels, insecureFlag, vars.Tail)
			} else {
				podName = s[0]
				return logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, rotatedFlag, allContainersFlag, logLevels, insecureFlag, vars.Tail)
			}
		}
		if len(args) == 2 {
			if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
				if containerName != "" {
					return fmt.Errorf("only one of -c or an inline [CONTAINER] arg is allowed")
				} else {
					podName = s[1]
					if podName == "" {
						return fmt.Errorf("arguments in resource/name form must have a single resource and name")
					}
					containerName = args[1]
					return logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, rotatedFlag, allContainersFlag, logLevels, insecureFlag, vars.Tail)
				}
			} else {
				if containerName != "" {
					return fmt.Errorf("only one of -c or an inline [CONTAINER] arg is allowed")
				} else {
					podName = args[0]
					containerName = args[1]
					return logsPods(vars.MustGatherRootPath, vars.Namespace, podName, containerName, previousFlag, rotatedFlag, allContainersFlag, logLevels, insecureFlag, vars.Tail)
				}
			}
		}
		return nil
	},
}

func init() {
	Logs.PersistentFlags().StringVarP(&vars.Container, "container", "c", "", "Print the logs of this container")
	Logs.PersistentFlags().BoolVar(&vars.InsecureLogs, "insecure", false, "")
	Logs.PersistentFlags().BoolVarP(&vars.Previous, "previous", "p", false, "Print the logs for the previous instance of the container in a pod if it exists.")
	Logs.PersistentFlags().BoolVarP(&vars.Rotated, "rotated", "r", false, "Print the logs for the rotated instance of the container in a pod if it exists.")
	Logs.PersistentFlags().BoolVarP(&vars.AllContainers, "all-containers", "", false, "Get all containers' logs in the pod(s).")
	Logs.PersistentFlags().Int64Var(&vars.Tail, "tail", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines.")
	Logs.Flags().StringVarP(&LogLevel, "log-level", "l", "", "Filter logs by level (info|error|worning), you can filter for more concatenating them comma separated.")
}
