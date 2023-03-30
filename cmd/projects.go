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

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
)

func projectsDefault(omcConfigFile string, projDefault string) {
	var namespaces []string
	_namespaces, _ := ioutil.ReadDir(vars.MustGatherRootPath + "/namespaces/")
	for _, f := range _namespaces {
		namespaces = append(namespaces, f.Name())
	}
	file, _ := ioutil.ReadFile(omcConfigFile)
	omcConfigJson := types.Config{}
	_ = json.Unmarshal([]byte(file), &omcConfigJson)

	config := types.Config{}

	var contexts []types.Context
	var NewContexts []types.Context
	contexts = omcConfigJson.Contexts
	for _, c := range contexts {
		if c.Current == "*" {
			if projDefault == "" {
				fmt.Println("Using project \"" + c.Project + "\" on must-gather \"" + c.Path + "\".")
				NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: c.Current, Project: c.Project})
			} else {
				if !helpers.StringInSlice(projDefault, namespaces) {
					fmt.Fprintln(os.Stderr, "Error: namespace "+projDefault+" does not exists in must-gather \""+c.Path+"\".")
					os.Exit(1)
				}
				NewContexts = append(NewContexts, types.Context{Id: c.Id, Path: c.Path, Current: c.Current, Project: projDefault})
				fmt.Println("Now using project \"" + projDefault + "\" on must-gather \"" + c.Path + "\".")
			}
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

var ProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Switch to another project",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			fmt.Fprintln(os.Stderr, "Expect zero arguemnt, found: ", len(args))
			os.Exit(1)
		}
		if len(args) == 0 {
			fmt.Println("You have access to the following projects and can switch between them with ' project <projectname>':")
			fmt.Println("")
			_namespaces, _ := ioutil.ReadDir(vars.MustGatherRootPath + "/namespaces/")
			for _, f := range _namespaces {
				if f.Name() == vars.Namespace {
					fmt.Println("  * ", f.Name())
				} else {
					fmt.Println("    ", f.Name())
				}
			}
			fmt.Println("")
			fmt.Println("Using project \"" + vars.Namespace + "\" on must-gather \"" + vars.MustGatherRootPath + "\".")
		}
	},
}
