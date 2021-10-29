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
	"strings"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var namespace string
var id string
var output string
var currentContextPath string
var defaultConfigNamespace string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{ // FLOW 4
	Use: "omc",
	Run: func(cmd *cobra.Command, args []string) { fmt.Println("Hello from omc CLI. :]") },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	//fmt.Println("inside init") //FLOW 0
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file to use (default is $HOME/.omc.json).")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "If present, list the requested object(s) for a specific namespace.")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// fmt.Println("inside initConfig") FLOW 1
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		// Search config in home directory with name ".omc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".omc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		omcConfigJson := models.Config{}
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())
		_ = json.Unmarshal([]byte(file), &omcConfigJson)
		var contexts []models.Context
		contexts = omcConfigJson.Contexts
		for _, context := range contexts {
			if context.Current == "*" {
				currentContextPath = context.Path
				defaultConfigNamespace = context.Project
				break
			}
		}
		if currentContextPath == "" {
			fmt.Println("There are no must-gather resources defined.")
			os.Exit(1)
		}
		exist, _ := helpers.Exists(currentContextPath + "/namespaces")
		if !exist {
			files, err := ioutil.ReadDir(currentContextPath)
			if err != nil {
				log.Fatal(err)
			}
			quayDir := ""
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "quay") {
					quayDir = f.Name()
					currentContextPath = currentContextPath + "/" + quayDir
					break
				}
			}
			if quayDir == "" {
				fmt.Println("Some error occurred, wrong must-gather file composition")
				os.Exit(1)
			}
		}
	} else {
		homePath, _ := os.UserHomeDir()
		helpers.CreateConfigFile(homePath)
		// TODO create the config file
	}
}
