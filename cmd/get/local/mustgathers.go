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
package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"omc/types"
	"os"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// contextsCmd represents the mg command
var MustGather = &cobra.Command{
	Use: "mg",
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())
		omcConfigJson := types.Config{}
		_ = json.Unmarshal([]byte(file), &omcConfigJson)

		var data [][]string
		var emptyData [][]string
		headers := []string{"current", "id", "path", "namespace"}
		var mg []types.Context
		mg = omcConfigJson.Contexts
		for _, context := range mg {
			_list := []string{context.Current, context.Id, context.Path, context.Project}
			data = append(data, _list)
		}
		if reflect.DeepEqual(data, emptyData) {
			fmt.Println("There are no must-gather resources defined.")
			os.Exit(1)
		} else {
			helpers.PrintTable(headers, data)
		}
	},
}
