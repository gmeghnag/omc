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

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

type ClusterRoleBindingsItems struct {
	ApiVersion string                      `json:"apiVersion"`
	Items      []rbacv1.ClusterRoleBinding `json:"items"`
}

func getClusterRoleBindings(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {
	clusterrolebindingsFolderPath := currentContextPath + "/cluster-scoped-resources/rbac.authorization.k8s.io/clusterrolebindings/"
	_clusterrolebindings, _ := ioutil.ReadDir(clusterrolebindingsFolderPath)

	_headers := []string{"name", "role", "age", "users", "groups", "serviceaccounts"}
	var data [][]string

	_ClusterRoleBindingsList := ClusterRoleBindingsItems{ApiVersion: "v1"}
	for _, f := range _clusterrolebindings {
		clusterrolebindingYamlPath := clusterrolebindingsFolderPath + f.Name()
		_file := helpers.ReadYaml(clusterrolebindingYamlPath)
		ClusterRoleBinding := rbacv1.ClusterRoleBinding{}
		if err := yaml.Unmarshal([]byte(_file), &ClusterRoleBinding); err != nil {
			fmt.Println("Error when trying to unmarshal file: " + clusterrolebindingYamlPath)
			os.Exit(1)
		}

		labels := helpers.ExtractLabels(ClusterRoleBinding.GetLabels())
		if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
			continue
		}
		if resourceName != "" && resourceName != ClusterRoleBinding.Name {
			continue
		}

		if outputFlag == "name" {
			_ClusterRoleBindingsList.Items = append(_ClusterRoleBindingsList.Items, ClusterRoleBinding)
			fmt.Println("clusterrolebinding/" + ClusterRoleBinding.Name)
			continue
		}

		if outputFlag == "yaml" {
			_ClusterRoleBindingsList.Items = append(_ClusterRoleBindingsList.Items, ClusterRoleBinding)
			continue
		}

		if outputFlag == "json" {
			_ClusterRoleBindingsList.Items = append(_ClusterRoleBindingsList.Items, ClusterRoleBinding)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_ClusterRoleBindingsList.Items = append(_ClusterRoleBindingsList.Items, ClusterRoleBinding)
			continue
		}

		role := ClusterRoleBinding.RoleRef.Kind + "/" + ClusterRoleBinding.RoleRef.Name
		users := []string{}
		groups := []string{}
		serviceaccounts := []string{}
		for _, subject := range ClusterRoleBinding.Subjects {
			if subject.Kind == "ServiceAccount" {
				serviceaccounts = append(serviceaccounts, subject.Namespace+subject.Name)
			}
			if subject.Kind == "User" {
				users = append(users, subject.Namespace+subject.Name)
			}
			if subject.Kind == "Group" {
				groups = append(groups, subject.Namespace+subject.Name)
			}
		}

		age := helpers.GetAge(clusterrolebindingYamlPath, ClusterRoleBinding.GetCreationTimestamp())

		_list := []string{ClusterRoleBinding.Name, role, age, strings.Join(users, ", "), strings.Join(groups, ", "), strings.Join(serviceaccounts, ", ")}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 3, _list)
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
		resource = _ClusterRoleBindingsList.Items[0]
	} else {
		resource = _ClusterRoleBindingsList
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

var ClusterRoleBinding = &cobra.Command{
	Use:     "clusterrolebinding",
	Aliases: []string{"clusterrolebinding", "clusterrolebindings"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getClusterRoleBindings(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}
