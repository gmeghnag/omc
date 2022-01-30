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
package alert

import (
	"os"
	"strings"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

// alertCmd represents the alert command
var AlertCmd = &cobra.Command{
	Use:     "alert",
	Aliases: []string{"alerts"},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	if len(os.Args) > 2 && (os.Args[1] == "alert" || os.Args[1] == "alerts") {
		if strings.Contains(os.Args[2], "/") {
			seg := strings.Split(os.Args[2], "/")
			resource, name := seg[0], seg[1]
			os.Args = append([]string{os.Args[0], "alert", resource, name}, os.Args[3:]...)
		}
	}
	AlertCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml")
	AlertCmd.AddCommand(
		GroupSubCmd,
		RuleSubCmd,
	)
}
