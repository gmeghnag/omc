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
package certificate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"omc/cmd/helpers"
	"reflect"
	"omc/vars"
	"os"
	"strings"

	v1 "k8s.io/api/certificates/v1"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)


func getCertificateSigningRequests(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, outputFlag string, showLabels bool, jsonPathTemplate string) bool {

	certificatesigningrequestsFolderPath := currentContextPath + "/cluster-scoped-resources/certificates.k8s.io/certificatesigningrequests/"
	_certificatesigningrequests, _ := ioutil.ReadDir(certificatesigningrequestsFolderPath)

	_headers := []string{"name", "age", "signername", "requestor", "condition"}
	var data [][]string

	_CertificateSigningRequestsList := v1.CertificateSigningRequestList{}
	for _, f := range _certificatesigningrequests {
		certificatesigningrequestYamlPath := certificatesigningrequestsFolderPath + f.Name()
		_file := helpers.ReadYaml(certificatesigningrequestYamlPath)
		CertificateSigningRequest := v1.CertificateSigningRequest{}
		if err := yaml.Unmarshal([]byte(_file), &CertificateSigningRequest); err != nil {
			fmt.Println("Error when trying to unmarshall file: " + certificatesigningrequestYamlPath)
			os.Exit(1)
		}

		if resourceName != "" && resourceName != CertificateSigningRequest.Name {
			continue
		}

		if outputFlag == "yaml" {
			_CertificateSigningRequestsList.Items = append(_CertificateSigningRequestsList.Items, CertificateSigningRequest)
			continue
		}

		if outputFlag == "json" {
			_CertificateSigningRequestsList.Items = append(_CertificateSigningRequestsList.Items, CertificateSigningRequest)
			continue
		}

		if strings.HasPrefix(outputFlag, "jsonpath=") {
			_CertificateSigningRequestsList.Items = append(_CertificateSigningRequestsList.Items, CertificateSigningRequest)
			continue
		}
		//Name
		certificatesigningrequestName := CertificateSigningRequest.Name
		age := helpers.GetAge(certificatesigningrequestYamlPath, CertificateSigningRequest.GetCreationTimestamp())
		
		//signername
		signername:= CertificateSigningRequest.Spec.SignerName
		//requestor
		requestor:= CertificateSigningRequest.Spec.Username

		//condition
		condition := "Unknown"
		if reflect.DeepEqual(CertificateSigningRequest.Status, v1.CertificateSigningRequestStatus{}) {
			condition = "Pending"
		} else {
		    for _, c := range CertificateSigningRequest.Status.Conditions {
		    	//Approved
		    	if c.Type == "Approved" {
		    		condition = "Approved,Issued"
		    		break
		    	}
		    	//Denied
		    	if c.Type == "Denied" {
		    		condition = "Denied"
		    		break
		    	}
		    	//Failed
		    	if c.Type == "Failed" {
		    		condition = "Failed"
		    		break
				}
				//Pending
		    	if c.Type == "Pending" {
		    		condition = "Pending"
		    		break
				}
		    }
	    }
		labels := helpers.ExtractLabels(CertificateSigningRequest.GetLabels())
		_list := []string{certificatesigningrequestName, age, signername, requestor, condition}
		data = helpers.GetData(data, true, showLabels, labels, outputFlag, 5, _list)
	}

	var headers []string
	if outputFlag == "" {
		headers = _headers[0:5] // -A
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
		resource = _CertificateSigningRequestsList.Items[0]
	} else {
		resource = _CertificateSigningRequestsList
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

var CertificateSigningRequest = &cobra.Command{
	Use:     "certificatesigningrequest",
	Aliases: []string{"certificatesigningrequest", "csr"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		getCertificateSigningRequests(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, vars.OutputStringVar, vars.ShowLabelsBoolVar, jsonPathTemplate)
	},
}
