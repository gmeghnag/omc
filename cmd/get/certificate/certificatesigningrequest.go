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
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/certificates/v1"

	"sigs.k8s.io/yaml"
)

func GetCertificateSigningRequests(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, out *[]v1.CertificateSigningRequest) {

	certificatesigningrequestsFolderPath := currentContextPath + "/cluster-scoped-resources/certificates.k8s.io/certificatesigningrequests/"
	_certificatesigningrequests, _ := ioutil.ReadDir(certificatesigningrequestsFolderPath)

	for _, f := range _certificatesigningrequests {
		certificatesigningrequestYamlPath := certificatesigningrequestsFolderPath + f.Name()
		_file := helpers.ReadYaml(certificatesigningrequestYamlPath)
		CertificateSigningRequest := v1.CertificateSigningRequest{}
		if err := yaml.Unmarshal([]byte(_file), &CertificateSigningRequest); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+certificatesigningrequestYamlPath)
			os.Exit(1)
		}

		if vars.LabelSelectorStringVar != "" {
			labels := helpers.ExtractLabels(CertificateSigningRequest.GetLabels())
			if !helpers.MatchLabels(labels, vars.LabelSelectorStringVar) {
				continue
			}
		}
		*out = append(*out, CertificateSigningRequest)
	}
}

var CertificateSigningRequest = &cobra.Command{
	Use:     "certificatesigningrequest",
	Aliases: []string{"certificatesigningrequests", "csr", "certificates.k8s.io"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		resourceName := ""
		if len(args) == 1 {
			resourceName = args[0]
		}
		var resources []v1.CertificateSigningRequest
		GetCertificateSigningRequests(vars.MustGatherRootPath, vars.Namespace, resourceName, vars.AllNamespaceBoolVar, &resources)
		if len(resources) == 0 {
			fmt.Fprintln(os.Stderr, "No resources found.")
			os.Exit(0)
		}
		_headers := []string{"name", "age", "signername", "requestor", "condition"}
		var data [][]string
		for _, CertificateSigningRequest := range resources {
			labels := helpers.ExtractLabels(CertificateSigningRequest.GetLabels())

			certificatesigningrequestName := CertificateSigningRequest.Name
			age := helpers.GetAge(vars.MustGatherRootPath+"/cluster-scoped-resources/certificates.k8s.io/certificatesigningrequests/", CertificateSigningRequest.GetCreationTimestamp())

			//signername
			signername := CertificateSigningRequest.Spec.SignerName
			//requestor
			requestor := CertificateSigningRequest.Spec.Username

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
			_list := []string{certificatesigningrequestName, age, signername, requestor, condition}
			data = helpers.GetData(data, true, vars.ShowLabelsBoolVar, labels, vars.OutputStringVar, 5, _list)
		}
		// ugly hack to get single item out of the slice
		//  TODO: handle this is helpets.PrintOutput
		var resourceSliceOrSingle interface{}
		if resourceName == "" {
			resourceSliceOrSingle = v1.CertificateSigningRequestList{Items: resources}
		} else {
			resourceSliceOrSingle = resources[0]
		}
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		helpers.PrintOutput(resourceSliceOrSingle, 4, vars.OutputStringVar, resourceName, vars.AllNamespaceBoolVar, vars.ShowLabelsBoolVar, _headers, data, jsonPathTemplate)
	},
}
