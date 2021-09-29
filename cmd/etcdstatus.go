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
	"io/ioutil"
	"log"
	"omc/cmd/helpers"
	"os"
	"strings"

	etcd "omc/cmd/etcd"

	"github.com/spf13/cobra"
)

func etcdStatusCommand() {
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
		var QuayString string
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "quay") {
				QuayString = f.Name()
				currentContextPath = currentContextPath + "/" + QuayString
				break
			}
		}
		if QuayString == "" {
			fmt.Println("Some error occurred, wrong must-gather file composition")
			os.Exit(1)
		}
	}
	etcdFolderPath := currentContextPath + "/etcd_info/"
	etcd.EndpointStatus(etcdFolderPath)

}

// etcdCmd represents the etcd command
var etcdStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Etcd status",
	Run: func(cmd *cobra.Command, args []string) {
		etcdStatusCommand()
	},
}

func init() {
	etcdCmd.AddCommand(etcdStatusCmd)
}
