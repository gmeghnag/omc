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
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// contextsCmd represents the mg command
var contextsCmd = &cobra.Command{
	Use:   "mg",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())
		omcConfigJson := models.Config{}
		_ = json.Unmarshal([]byte(file), &omcConfigJson)

		var data [][]string
		headers := []string{"current", "id", "path", "namespace"}
		var mg []models.Context
		mg = omcConfigJson.Contexts
		for _, context := range mg {
			_list := []string{context.Current, context.Id, context.Path, context.Project}
			data = append(data, _list)
		}
		helpers.PrintTable(headers, data)

	},
}

func init() {
	getCmd.AddCommand(contextsCmd)

}
