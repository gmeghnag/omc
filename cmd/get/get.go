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
package get

import (
	"fmt"
	"omc/cmd/get/core"
	"omc/vars"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var outputStringVar string
var allNamespaceBoolVar, showLabelsBoolVar bool

var GetCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called", args)
	},
}

func init() {
	if os.Args[1] == "get" && len(os.Args) > 2 {
		if strings.Contains(os.Args[2], "/") {
			seg := strings.Split(os.Args[2], "/")
			resource, name := seg[0], seg[1]
			os.Args = append([]string{os.Args[0], "get", resource, name}, os.Args[3:]...)
		}
	}

	GetCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	GetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	GetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	//getCmd.PersistentFlags().StringVarP(&selector, "selector", "l", "", "elector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	GetCmd.AddCommand(
		Build,
		BuildConfig,
		ClusterOperator,
		ClusterVersion,
		core.ConfigMap,
		DaemonSet,
		Deployment,
		DeploymentConfig,
		core.Event,
		ImageStream,
		Job,
		Node,
		core.Pod,
	)
}
