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
	"log"
	"os"
	"strings"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DeleteAll bool

func DeleteContext(path string, omcConfigFile string, idFlag string) {
	// read json omcConfigFile

	if DeleteAll {
		file, _ := json.MarshalIndent(types.Config{}, "", " ")
		_ = ioutil.WriteFile(omcConfigFile, file, 0644)
		os.Exit(0)
	}

	file, _ := ioutil.ReadFile(omcConfigFile)
	omcConfigJson := types.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	config := types.Config{}

	var contexts []types.Context
	var NewContexts []types.Context
	contexts = omcConfigJson.Contexts
	for _, c := range contexts {
		if c.Id == idFlag || c.Path == path {
			continue
		} else {
			NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: c.Current, Project: c.Project})
		}
	}

	config.Contexts = NewContexts
	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		log.Fatal("Json Marshal failed")
	}
	_ = ioutil.WriteFile(omcConfigFile, file, 0644)

}

// useCmd represents the use command
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete mg from the saved.",
	Run: func(cmd *cobra.Command, args []string) {
		idFlag, _ := cmd.Flags().GetString("id")
		path := ""
		if len(args) > 1 {
			fmt.Println("Expect one arguemnt, found: ", len(args))
			os.Exit(1)
		}
		if len(args) == 1 {
			path = args[0]
			if strings.HasSuffix(path, "/") {
				path = strings.TrimRight(path, "/")
			}
		}

		DeleteContext(path, viper.ConfigFileUsed(), idFlag)
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&vars.Id, "id", "i", "", "Id string for the must-gather. If two must-gather has the same id the first one will be used.")
	DeleteCmd.Flags().BoolVarP(&DeleteAll, "all", "a", false, "Delete all referenced must-gathers.")
}
