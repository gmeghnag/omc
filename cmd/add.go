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
	"log"
	"omc/cmd/helpers"
	"omc/models"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addContext(path string, omcConfigFile string) {
	if !filepath.IsAbs(path) {
		log.Fatal(": '", path, "' is not an absolute path.")
	}
	// read json omcConfigFile
	file, _ := ioutil.ReadFile(omcConfigFile)
	omcConfigJson := models.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	// create new context id
	ctxId := helpers.RandString(8)
	config := models.Config{}

	var contexts []models.Context
	var NewContexts []models.Context
	contexts = omcConfigJson.Contexts
	for _, c := range contexts {
		c.Current = ""
		NewContexts = append(NewContexts, c)
	}
	NewContexts = append(NewContexts, models.Context{Id: ctxId, Path: path, Current: "*", Project: "default"})

	config.Contexts = NewContexts
	config.Id = ctxId
	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		log.Fatal("Json Marshal failed")
	}
	_ = ioutil.WriteFile(omcConfigFile, file, 0644)

}

// useCmd represents the use command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Expect one arguemnt, found: ", len(args))
		}
		path := args[0]
		addContext(path, viper.ConfigFileUsed())
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// useCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// useCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
