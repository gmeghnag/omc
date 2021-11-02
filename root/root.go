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
package root

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"omc/cmd"
	"omc/cmd/describe"
	"omc/cmd/etcd"
	"omc/cmd/get"
	"omc/cmd/helpers"
	"omc/cmd/logs"
	"omc/types"
	"omc/vars"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{ // FLOW 4
	Use: "omc",
	Run: func(cmd *cobra.Command, args []string) { fmt.Println("Hello from omc CLI. :]") },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	//fmt.Println("inside init") //FLOW 0
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&vars.CfgFile, "config", "", "Config file to use (default is $HOME/.omc.json).")
	RootCmd.PersistentFlags().StringVarP(&vars.Namespace, "namespace", "n", "", "If present, list the requested object(s) for a specific namespace.")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.AddCommand(
		cmd.VersionCmd,
		cmd.DeleteCmd,
		cmd.ProjectCmd,
		cmd.UseCmd,
		get.GetCmd,
		describe.DescribeCmd,
		etcd.Etcd,
		logs.Logs,
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// fmt.Println("inside initConfig") FLOW 1
	if vars.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(vars.CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		exist, _ := helpers.Exists(home + "/.omc.json")
		if !exist {
			config := types.Config{}
			file, _ := json.MarshalIndent(config, "", " ")
			_ = ioutil.WriteFile(vars.CfgFile, file, 0644)
		}
		// Search config in home directory with name ".omc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".omc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		omcConfigJson := types.Config{}
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())
		_ = json.Unmarshal([]byte(file), &omcConfigJson)
		var contexts []types.Context
		contexts = omcConfigJson.Contexts
		for _, context := range contexts {
			if context.Current == "*" {
				vars.MustGatherRootPath = context.Path
				if vars.Namespace == "" {
					vars.Namespace = context.Project
				}
				break
			}
		}
		if vars.MustGatherRootPath != "" {
			exist, _ := helpers.Exists(vars.MustGatherRootPath + "/namespaces")
			if !exist {
				files, err := ioutil.ReadDir(vars.MustGatherRootPath)
				if err != nil {
					log.Fatal(err)
				}
				baseDir := ""
				for _, f := range files {
					if f.IsDir() {
						baseDir = f.Name()
						vars.MustGatherRootPath = vars.MustGatherRootPath + "/" + baseDir
						break
					}
				}
				if baseDir == "" {
					fmt.Println("Some error occurred, wrong must-gather file composition")
					os.Exit(1)
				}
			}
		}
	} else {
		homePath, _ := os.UserHomeDir()
		helpers.CreateConfigFile(homePath)
		// TODO create the config file
	}
}
