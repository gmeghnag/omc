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
package resources

import (
	"fmt"
	"os"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

// alertCmd represents the alert command
var ApiResourcesCmd = &cobra.Command{
	Use: "api-resources",
	Run: func(cmd *cobra.Command, args []string) {
		err := getResources(vars.OutputStringVar)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	//if len(os.Args) > 2 && (os.Args[1] == "api-resources") {
	//	fmt.Println("error: unexpected arguments:", os.Args[2:])
	//	os.Exit(1)
	//}
	ApiResourcesCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide")
}
