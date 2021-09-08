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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func versionOmc() {
	fmt.Println("omc version 1.0 (https://github.com/gmeghnag/omc)")
}

// useCmd represents the use command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print omc version",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("Unexpected argument: " + args[0] + ".")
			os.Exit(1)
		}

		versionOmc()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
