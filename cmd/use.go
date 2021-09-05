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
	"omc/cmd/helpers"
	"omc/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func useContext(path string, omcConfigFile string, idFlag string) {
	if path != "" {
		if !filepath.IsAbs(path) {
			fmt.Println(": '", path, "' is not an absolute path.")
			os.Exit(1)
		}
	}
	// read json omcConfigFile
	file, _ := ioutil.ReadFile(omcConfigFile)
	omcConfigJson := models.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	config := models.Config{}

	var contexts []models.Context
	var NewContexts []models.Context
	contexts = omcConfigJson.Contexts
	var found bool
	var ctxId string
	for _, c := range contexts {
		if c.Id == idFlag || c.Path == path {
			NewContexts = append(NewContexts, models.Context{Id: c.Id, Path: c.Path, Current: "*", Project: c.Project})
			found = true
		} else {
			NewContexts = append(NewContexts, models.Context{Id: c.Id, Path: c.Path, Current: "", Project: c.Project})
		}
	}
	if !found {
		if idFlag != "" {
			NewContexts = append(NewContexts, models.Context{Id: idFlag, Path: path, Current: "*", Project: "default"})
		} else {
			ctxId = helpers.RandString(8)
			NewContexts = append(NewContexts, models.Context{Id: ctxId, Path: path, Current: "*", Project: "default"})
		}

	}

	config.Contexts = NewContexts
	config.Id = idFlag
	if !found {
		if idFlag != "" {
			config.Id = idFlag
		} else {
			config.Id = ctxId
		}
	}
	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		log.Fatal("Json Marshal failed")
	}
	_ = ioutil.WriteFile(omcConfigFile, file, 0644)

}

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		useContext(path, viper.ConfigFileUsed(), idFlag)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().StringVarP(&id, "id", "i", "", "Id string for the must-gather. If two must-gather has the same id the first one will be used.")
}
