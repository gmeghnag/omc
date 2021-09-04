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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/models"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var output string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		allNamespacesFlag, _ := cmd.Flags().GetBool("all-namespaces")
		outputFlag, _ := cmd.Flags().GetString("output")
		namespace, _ := rootCmd.PersistentFlags().GetString("namespace")
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())

		omcConfigJson := models.Config{}
		_ = json.Unmarshal([]byte(file), &omcConfigJson)
		var CurrentContextPath string
		var DefaultConfigNamespace string
		var contexts []models.Context
		contexts = omcConfigJson.Contexts
		for _, context := range contexts {
			if context.Current == "*" {
				CurrentContextPath = context.Path
				DefaultConfigNamespace = context.Project
				break
			}
		}
		jsonPathTemplate := ""
		if strings.HasPrefix(outputFlag, "jsonpath=") {
			s := strings.Split(outputFlag, "=")
			if len(s) < 2 || s[1] == "" {
				fmt.Println("error: template format specified but no template given")
				os.Exit(1)
			}
			jsonPathTemplate = s[1]
		}

		if len(args) == 1 {
			argument := args[0]
			if argument == "pod" || argument == "pods" {
				getPods(CurrentContextPath, DefaultConfigNamespace, "", allNamespacesFlag, outputFlag, jsonPathTemplate)
			}
			if strings.HasPrefix(argument, "pod/") || strings.HasPrefix(argument, "pods/") {
				s := strings.Split(argument, "/")
				getPods(CurrentContextPath, DefaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, jsonPathTemplate)
			}
			if argument == "node" || argument == "nodes" {
				getNodes(CurrentContextPath, DefaultConfigNamespace, "", allNamespacesFlag, outputFlag, jsonPathTemplate)
			}
			if strings.HasPrefix(argument, "node/") || strings.HasPrefix(argument, "nodes/") {
				s := strings.Split(argument, "/")
				getNodes(CurrentContextPath, DefaultConfigNamespace, s[1], allNamespacesFlag, outputFlag, jsonPathTemplate)
			}
		} else {
			fmt.Println("No resources found in " + namespace + " namespace")
		}
	},
}

func init() {
	//fmt.Println("inside get init")

	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().BoolP("all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	getCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
