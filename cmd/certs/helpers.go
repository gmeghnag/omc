package certs

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gmeghnag/omc/cmd/helpers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type ResourcesItems struct {
	Kind       string                       `json:"kind"`
	ApiVersion string                       `json:"apiVersion"`
	Items      []*unstructured.Unstructured `json:"items"`
}

func GetSecrets(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, out *[]*unstructured.Unstructured) {
	var namespaces []string
	if allNamespacesFlag == true {
		namespace = "all"
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, namespace)
	}

	for _, _namespace := range namespaces {
		var _Items ResourcesItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/secrets.yaml")
		if err != nil && !allNamespacesFlag {
			continue
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+CurrentNamespacePath+"/core/secrets.yaml")
			os.Exit(1)
		}

		for _, Secret := range _Items.Items {
			*out = append(*out, Secret)
		}
	}
}

func GetConfigMaps(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, out *[]*unstructured.Unstructured) {
	var namespaces []string
	if allNamespacesFlag == true {
		namespace = "all"
		_namespaces, _ := ioutil.ReadDir(currentContextPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, namespace)
	}

	for _, _namespace := range namespaces {
		var _Items ResourcesItems
		CurrentNamespacePath := currentContextPath + "/namespaces/" + _namespace
		_file, err := ioutil.ReadFile(CurrentNamespacePath + "/core/configmaps.yaml")
		if err != nil && !allNamespacesFlag {
			continue
		}
		if err := yaml.Unmarshal([]byte(_file), &_Items); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+CurrentNamespacePath+"/core/configmaps.yaml")
			os.Exit(1)
		}
		for _, ConfigMap := range _Items.Items {
			*out = append(*out, ConfigMap)
		}
	}
}

func GetCertificateSigningRequests(currentContextPath string, namespace string, resourceName string, allNamespacesFlag bool, out *[]unstructured.Unstructured) {

	certificatesigningrequestsFolderPath := currentContextPath + "/cluster-scoped-resources/certificates.k8s.io/certificatesigningrequests/"
	_certificatesigningrequests, _ := ioutil.ReadDir(certificatesigningrequestsFolderPath)

	for _, f := range _certificatesigningrequests {
		certificatesigningrequestYamlPath := certificatesigningrequestsFolderPath + f.Name()
		_file := helpers.ReadYaml(certificatesigningrequestYamlPath)
		CertificateSigningRequest := unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(_file), &CertificateSigningRequest); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+certificatesigningrequestYamlPath)
			os.Exit(1)
		}
		*out = append(*out, CertificateSigningRequest)
	}
}
