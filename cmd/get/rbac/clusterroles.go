/*
Copyright © 2022 Bram Verschueren <bverschueren@redhat.com>

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
package rbac

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
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

type ClusterRolesItems struct {
	ApiVersion string               `json:"apiVersion"`
	Items      []rbacv1.ClusterRole `json:"items"`
}

func getClusterRoles(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	clusterrolesFolderPath := currentContextPath + "/cluster-scoped-resources/rbac.authorization.k8s.io/clusterroles/"
	_clusterroles, _ := ioutil.ReadDir(clusterrolesFolderPath)

	_headers := []string{"name", "created at"}
	var data [][]string

	_ClusterRolesList := ClusterRolesItems{ApiVersion: "v1"}
	for _, f := range _clusterroles {
		clusterroleYamlPath := clusterrolesFolderPath + f.Name()
		_file := helpers.ReadYaml(clusterroleYamlPath)
		ClusterRole := rbacv1.ClusterRole{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterRole); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + clusterroleYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ClusterRole.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterRole.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterRolesList.Items = append(_ClusterRolesList.Items, ClusterRole)
			fmt.Println("clusterrole/" + ClusterRole.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ClusterRolesList.Items = append(_ClusterRolesList.Items, ClusterRole)
			continue
		}

		if outputFlag == "json" {
			_ClusterRolesList.Items = append(_ClusterRolesList.Items, ClusterRole)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ClusterRolesList.Items = append(_ClusterRolesList.Items, ClusterRole)
			continue
		}

		creationTimestamp := ClusterRole.GetCreationTimestamp()
		createdAt := creationTimestamp.UTC().Format(time.RFC3339Nano)

		_list := []string{ClusterRole.Name, createdAt}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 2, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:2] // -A
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		headers = _headers // -A -o wide
		if showLabels {
			headers = append(headers, "labels")
		}
		helpers.PrintTable(headers, data)
		return false
	}
	var resource interface{}
	if resourceName != "" {
		resource = _ClusterRolesList.Items[0]
	} else {
		resource = _ClusterRolesList
	}
	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(resource)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		helpers.ExecuteJsonPath(resource, jsonPathTemplate)
	}
	return false
}

var ClusterRole = &cobra.Command{
	Use:     "clusterrole",
	Aliases: []string{"clusterrole", "clusterroles"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterRoles(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}
