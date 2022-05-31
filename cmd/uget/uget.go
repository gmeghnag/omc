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
package uget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
)

var outputStringVar, additionalColumnsPath, objectFilePath, kindStringVar string
var allNamespaceBoolVar, showLabelsBoolVar bool

var UGetCmd = &cobra.Command{
	Use:     "uget",
	Aliases: []string{"dget"},
	Run: func(cmd *cobra.Command, args []string) {
		if objectFilePath == "" {
			fmt.Println("The path for the object(s) to inspect needs to be defined with the flag --path")
			os.Exit(1)
		}
		UGet(objectFilePath, args)
		os.Exit(0)
	},
}

func init() {
	UGetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	UGetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	UGetCmd.PersistentFlags().StringVarP(&additionalColumnsPath, "columns", "c", "", "Costum columns file path.")
	UGetCmd.PersistentFlags().StringVarP(&objectFilePath, "path", "p", "", "Inspect object(s) path.")
	UGetCmd.PersistentFlags().StringVarP(&kindStringVar, "kind", "k", "", "kind(s) to filter on, single or multiple (comma speratad)")
	UGetCmd.PersistentFlags().StringVarP(&vars.LabelSelectorStringVar, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
}

func UGet(objectsPath string, objectsName []string) {
	defaultColumns := false
	returnObjects := UnstrctList{ApiVersion: "v1", Kind: "List"}
	if !PathExists(objectsPath) {
		fmt.Printf("Path %v does not exist.\n", objectsPath)
		os.Exit(1)
	}
	if additionalColumnsPath != "" && !PathExists(additionalColumnsPath) {
		fmt.Printf("File %v does not exist.\n", additionalColumnsPath)
		os.Exit(1)
	}
	var data [][]string
	var headers []string

	if additionalColumnsPath == "" {
		headers = append(headers, "kind", "name")
		defaultColumns = true
	}
	columnsByte, _ := ioutil.ReadFile(additionalColumnsPath)
	AdditionalColumnsStruct := AdditionalColumns{}
	if err := yaml.Unmarshal([]byte(columnsByte), &AdditionalColumnsStruct); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var JSONPaths []Column
	for _, column := range AdditionalColumnsStruct.Columns {
		headers = append(headers, column.Name)
		JSONPaths = append(JSONPaths, column)
	}

	var kindsArray []string
	if kindStringVar != "" {
		kindStringVar = strings.ToLower(kindStringVar)
		kindsArray = strings.Split(strings.TrimSuffix(kindStringVar, ","), ",")
	}

	var objectFiles []string
	fi, _ := os.Stat(objectsPath)
	if fi.Mode().IsDir() {
		resources, _ := ioutil.ReadDir(objectsPath)
		for _, f := range resources {
			if !f.IsDir() {
				objectFiles = append(objectFiles, strings.TrimSuffix(objectsPath, "/")+"/"+f.Name())
			}
		}
	} else {
		objectFiles = append(objectFiles, objectsPath)
	}

	for _, resourceYamlPath := range objectFiles {
		resourceByte, _ := ioutil.ReadFile(resourceYamlPath)
		unstruct := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(resourceByte), &unstruct); err != nil {
			fmt.Println("File:", resourceYamlPath, " does not contain a valid k8s object,", err.Error())
			os.Exit(1)
		}
		if unstruct.IsList() {
			unstructList := &unstructured.UnstructuredList{}
			err := yaml.Unmarshal([]byte(resourceByte), &unstructList)
			if err != nil {
				fmt.Println("File:", resourceYamlPath, " does not contain a valid k8s object,", err.Error())
				os.Exit(1)
			}
			for _, resource := range unstructList.Items {
				var resourceData []string
				resourceName := resource.GetName()
				labels := helpers.ExtractLabels(resource.GetLabels())
				if matchKind(kindsArray, resource.GetKind()) && helpers.MatchLabels(labels, vars.LabelSelectorStringVar) && (len(objectsName) == 0 || helpers.StringInSlice(resourceName, objectsName)) {
					if vars.OutputStringVar == "" {
						if defaultColumns {
							resourceData = append(resourceData, resource.GetKind(), resourceName)
						}
						for _, column := range AdditionalColumnsStruct.Columns {
							v := getFromJsonPath(resource.Object, toJsonPath(column.JSONPath))
							if column.Type == "date" {
								v = helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
							}
							resourceData = append(resourceData, v)
						}
						if vars.ShowLabelsBoolVar {
							resourceData = append(resourceData, helpers.ExtractLabels(resource.GetLabels()))
						}
					} else {
						returnObjects.Items = append(returnObjects.Items, resource)
					}
				}
				if len(resourceData) != 0 {
					data = append(data, resourceData)
				}
			}
		} else {
			// not a List
			resourceName := unstruct.GetName()
			labels := helpers.ExtractLabels(unstruct.GetLabels())
			if matchKind(kindsArray, unstruct.GetKind()) && helpers.MatchLabels(labels, vars.LabelSelectorStringVar) && (len(objectsName) == 0 || helpers.StringInSlice(resourceName, objectsName)) {
				if vars.OutputStringVar == "" {
					var resourceData []string
					if defaultColumns {
						resourceData = append(resourceData, unstruct.GetKind(), resourceName)
					}
					for _, jpath := range JSONPaths {
						v := getFromJsonPath(unstruct.Object, toJsonPath(jpath.JSONPath))
						if jpath.Type == "date" {
							v = helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
						}
						resourceData = append(resourceData, v)
					}
					if len(resourceData) != 0 {
						data = append(data, resourceData)
					}
				} else {
					returnObjects.Items = append(returnObjects.Items, *unstruct)
				}
			}
		}
	}
	if vars.OutputStringVar == "" {
		if len(data) == 0 {
			fmt.Println("No resources found.")
			os.Exit(1)
		} else {
			if vars.ShowLabelsBoolVar {
				headers = append(headers, "labels")
			}
			helpers.PrintTable(headers, data)
		}
	} else {
		if len(returnObjects.Items) == 0 {
			fmt.Println("No resources found.")
			os.Exit(1)
		}
		if vars.OutputStringVar == "json" {
			if len(returnObjects.Items) == 1 {
				j, _ := json.MarshalIndent(returnObjects.Items[0].Object, "", "  ")
				fmt.Println(string(j))
			} else {
				j, _ := json.MarshalIndent(returnObjects, "", "  ")
				fmt.Println(string(j))
			}
		} else if vars.OutputStringVar == "yaml" {
			if len(returnObjects.Items) == 1 {
				y, _ := yaml.Marshal(returnObjects.Items[0].Object)
				fmt.Println(string(y))
			} else {
				y, _ := yaml.Marshal(returnObjects)
				fmt.Println(string(y))
			}
		} else if strings.HasPrefix(vars.OutputStringVar, "jsonpath=") {
			jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
			if len(returnObjects.Items) == 1 {
				helpers.ExecuteJsonPath(returnObjects.Items[0].Object, jsonPathTemplate)
			} else {
				helpers.ExecuteJsonPath(returnObjects, jsonPathTemplate)
			}
		}
	}
}
