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
	"io"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/pkg/mustgather"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

var LogLevel string

// Flag binding targets. Cobra binds flags to variable addresses, so these
// stay as package vars. RunE copies them into an Options for Run.
var (
	containerFlag     string
	previousFlag      bool
	rotatedFlag       bool
	allContainersFlag bool
	insecureFlag      bool
	tailFlag          int64
)

// logsCmd represents the logs command
var Logs = &cobra.Command{
	Use:          "logs",
	Short:        "Print the logs for a container in a pod",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		namespace := vars.Namespace
		if namespaceFlag != "" {
			namespace = namespaceFlag
		}
		opts := Options{
			RootPath:      vars.MustGatherRootPath,
			RootPaths:     vars.MustGatherRootPaths,
			Namespace:     namespace,
			Container:     containerFlag,
			Previous:      previousFlag,
			Rotated:       rotatedFlag,
			AllContainers: allContainersFlag,
			Insecure:      insecureFlag,
			Tail:          tailFlag,
		}
		return Run(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts, args)
	},
}

func Run(stdout, stderr io.Writer, opts Options, args []string) error {
	if opts.RootPath == "" {
		return fmt.Errorf("there are no must-gather resources defined")
	}
	exist, _ := helpers.Exists(opts.RootPath + "/namespaces")
	if !exist {
		files, err := os.ReadDir(opts.RootPath)
		if err != nil {
			return err
		}
		var QuayString string
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "quay") {
				QuayString = f.Name()
				opts.RootPath = opts.RootPath + "/" + QuayString
				break
			}
		}
		if QuayString == "" {
			return fmt.Errorf("wrong must-gather file composition")
		}
	}
	podName := ""
	containerName := opts.Container
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
		} else {
			podName = s[0]
		}
	}
	if len(args) == 2 {
		if s := strings.Split(args[0], "/"); len(s) == 2 && (s[0] == "po" || s[0] == "pod" || s[0] == "pods") {
			if containerName != "" {
				return fmt.Errorf("only one of -c or an inline [CONTAINER] arg is allowed")
			}
			podName = s[1]
			if podName == "" {
				return fmt.Errorf("arguments in resource/name form must have a single resource and name")
			}
			containerName = args[1]
		} else {
			if containerName != "" {
				return fmt.Errorf("only one of -c or an inline [CONTAINER] arg is allowed")
			}
			podName = args[0]
			containerName = args[1]
		}
	}

	// In a grouped (multi-must-gather) context resolve the pod to the most
	// recent must-gather that contains it - the same capture get and describe
	// show - and read its logs only from there. If that capture has no logs for
	// the pod we show none rather than falling back to an older capture.
	rootPath := resolveLogsRoot(opts, podName)
	return logsPods(stdout, rootPath, opts.Namespace, podName, containerName, opts.Previous, opts.Rotated, opts.AllContainers, logLevels, opts.Insecure, opts.Tail)
}

// resolveLogsRoot returns the must-gather root a pod's logs must be read from.
// For a single-must-gather context it is just opts.RootPath. For a grouped
// context it is the most recent root that contains the pod; ResolveRoot falls
// back to the most recent root when the pod is found in none, letting logsPods
// report "not found" as usual.
func resolveLogsRoot(opts Options, podName string) string {
	if len(opts.RootPaths) <= 1 {
		return opts.RootPath
	}
	ns := opts.Namespace
	candidates := []string{
		"namespaces/" + ns + "/pods/" + podName + "/" + podName + ".yaml",
		"namespaces/" + ns + "/pods/" + podName,
		"namespaces/" + ns + "/core/pods/" + podName + ".yaml",
	}
	root, _ := mustgather.ResolveRoot(opts.RootPaths, candidates...)
	return root
}

func init() {
	Logs.PersistentFlags().StringVarP(&containerFlag, "container", "c", "", "Print the logs of this container")
	Logs.PersistentFlags().BoolVar(&insecureFlag, "insecure", false, "")
	Logs.PersistentFlags().BoolVarP(&previousFlag, "previous", "p", false, "Print the logs for the previous instance of the container in a pod if it exists.")
	Logs.PersistentFlags().BoolVarP(&rotatedFlag, "rotated", "r", false, "Print the logs for the rotated instance of the container in a pod if it exists.")
	Logs.PersistentFlags().BoolVarP(&allContainersFlag, "all-containers", "", false, "Get all containers' logs in the pod(s).")
	Logs.PersistentFlags().Int64Var(&tailFlag, "tail", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines.")
	Logs.Flags().StringVarP(&LogLevel, "log-level", "l", "", "Filter logs by level (info|error|worning), you can filter for more concatenating them comma separated.")
}

