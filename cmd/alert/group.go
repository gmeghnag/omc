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
package alert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func GetAlertGroups(resourcesNames []string, outputFlag string, groupFile string) {
	_headers := []string{"group", "filename", "age"}
	var data [][]string
	var filteredGroups []RuleGroup
	var _Alerts alerts
	_file, _ := ioutil.ReadFile(vars.MustGatherRootPath + "/monitoring/alerts.json")
	if err := yaml.Unmarshal([]byte(_file), &_Alerts); err != nil {
		fmt.Println("Error when trying to unmarshal file " + vars.MustGatherRootPath + "/monitoring/alerts.json")
		os.Exit(1)
	}

	for _, group := range _Alerts.Data.Groups {
		filename := group.File[strings.LastIndex(group.File, "/")+1:]
		if len(resourcesNames) != 0 && !helpers.StringInSlice(group.Name, resourcesNames) {
			continue
		}

		if groupFile != "" && filename != groupFile {
			continue
		}

		if outputFlag == "yaml" || outputFlag == "json" {
			filteredGroups = append(filteredGroups, group)
			continue
		}

		//fmt.Println(al.Name, filename)
		ResourceFile, _ := os.Stat(vars.MustGatherRootPath + "/monitoring/alerts.json")
		t2 := ResourceFile.ModTime()
		diffTime := t2.Sub(group.LastEvaluation).String()
		d, _ := time.ParseDuration(diffTime)
		lastEval := helpers.FormatDiffTime(d)
		_list := []string{group.Name, filename, lastEval}
		data = helpers.GetData(data, true, false, "", "", 3, _list)
	}

	var headers []string
	if outputFlag == "" || outputFlag == "wide" {
		headers = _headers[0:3]
		if len(data) == 0 {
			fmt.Println("No alertgroups found.")
		} else {
			helpers.PrintTable(headers, data)
		}
	}
	if outputFlag == "yaml" {
		_Alerts.Data.Groups = filteredGroups
		y, _ := yaml.Marshal(_Alerts)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		_Alerts.Data.Groups = filteredGroups
		j, _ := json.Marshal(_Alerts)
		fmt.Println(string(j))
	}

}

var GroupSubCmd = &cobra.Command{
	Use:     "group",
	Aliases: []string{"groups"},
	Run: func(cmd *cobra.Command, args []string) {
		resourcesNames := args
		GetAlertGroups(resourcesNames, vars.OutputStringVar, GroupFilename)
	},
}

func init() {
	GroupSubCmd.Flags().StringVarP(&GroupFilename, "filename", "f", "", "Filter the AlertGroup by filename.")
}
