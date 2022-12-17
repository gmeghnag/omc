package get

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"bytes"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/cmd/uget"
	"github.com/gmeghnag/omc/vars"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/yaml"
)

type AdditionalPrinterColumn struct {
	Description string `json:"id"`
	JsonPath    string `json:"path"`
	Name        string `json:"current"`
	Type        string `json:"project"`
}

var data [][]string
var headers []string

var returnObjects = uget.UnstrctList{ApiVersion: "v1", Kind: "List"}

func getGenericResourceFromCRD(crdName string, objectNames []string) bool {
	var crd *apiextensionsv1.CustomResourceDefinition
	crdsPath := vars.MustGatherRootPath + "/cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/"
	crdExists := false
	if strings.HasSuffix(crdName, ".config") {
		crdName = strings.Replace(crdName, ".config", ".config.openshift.io", -1)
	}
	_, err := Exists(crdsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	crds, _ := ioutil.ReadDir(crdsPath)
	for _, f := range crds {
		crdYamlPath := crdsPath + f.Name()
		crdByte, _ := ioutil.ReadFile(crdYamlPath)
		_crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal([]byte(crdByte), &_crd); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file", crdYamlPath)
			os.Exit(1)
		}
		if strings.ToLower(_crd.Name) == strings.ToLower(crdName) || strings.ToLower(_crd.Spec.Names.Plural) == strings.ToLower(crdName) || strings.ToLower(_crd.Spec.Names.Singular) == strings.ToLower(crdName) || helpers.StringInSlice(crdName, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == strings.ToLower(crdName) {
			crdExists = true
			crd = _crd
			break
		}
	}
	if !crdExists && vars.UseLocalCRDs {
		home, _ := os.UserHomeDir()
		crdsPath := home + "/.omc/customresourcedefinitions/"
		if strings.HasSuffix(crdName, ".config") {
			crdName = strings.Replace(crdName, ".config", ".config.openshift.io", -1)
		}
		_, err := Exists(crdsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		crds, _ := ioutil.ReadDir(crdsPath)
		for _, f := range crds {
			crdYamlPath := crdsPath + f.Name()
			crdByte, _ := ioutil.ReadFile(crdYamlPath)
			_crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal([]byte(crdByte), &_crd); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file", crdYamlPath)
				os.Exit(1)
			}
			if strings.ToLower(_crd.Name) == strings.ToLower(crdName) || strings.ToLower(_crd.Spec.Names.Plural) == strings.ToLower(crdName) || strings.ToLower(_crd.Spec.Names.Singular) == strings.ToLower(crdName) || helpers.StringInSlice(crdName, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == strings.ToLower(crdName) {
				crdExists = true
				crd = _crd
				break
			}
		}
	}
	if !crdExists {
		return false
	}
	if vars.AllNamespaceBoolVar && crd.Spec.Scope != "Cluster" {
		headers = append(headers, "Namespace")
	}
	headers = append(headers, "Name")
	if crd.Spec.Versions[0].AdditionalPrinterColumns == nil {
		headers = append(headers, "Age")
	}
	for _, column := range crd.Spec.Versions[0].AdditionalPrinterColumns {
		if column.Priority == 0 || vars.OutputStringVar == "wide" {
			headers = append(headers, column.Name)
		}
	}
	if crd.Spec.Scope == "Cluster" {
		resourcesPath := vars.MustGatherRootPath + "/cluster-scoped-resources/" + crd.Spec.Group + "/" + crd.Spec.Names.Plural + "/"
		bole, _ := Exists(resourcesPath)
		if !bole {
			resourcesPath = vars.MustGatherRootPath + "/cluster-scoped-resources/" + crd.Spec.Group + "/"
		}
		gatherObjects(resourcesPath, crd, objectNames)
	} else {
		if vars.AllNamespaceBoolVar {
			namespacesPath := vars.MustGatherRootPath + "/namespaces/"
			namespaces, _ := ioutil.ReadDir(namespacesPath)
			for _, ns := range namespaces {
				resourcesPath := vars.MustGatherRootPath + "/namespaces/" + ns.Name() + "/" + crd.Spec.Group + "/" + crd.Spec.Names.Plural + "/"
				bole, _ := Exists(resourcesPath)
				if !bole {
					resourcesPath = vars.MustGatherRootPath + "/namespaces/" + ns.Name() + "/" + crd.Spec.Group + "/"
				}
				gatherObjects(resourcesPath, crd, objectNames)
			}
		} else {
			resourcesPath := vars.MustGatherRootPath + "/namespaces/" + vars.Namespace + "/" + crd.Spec.Group + "/" + crd.Spec.Names.Plural + "/"
			bole, _ := Exists(resourcesPath)
			if !bole {
				resourcesPath = vars.MustGatherRootPath + "/namespaces/" + vars.Namespace + "/" + crd.Spec.Group + "/"
			}
			gatherObjects(resourcesPath, crd, objectNames)
		}
	}
	if vars.OutputStringVar == "" || vars.OutputStringVar == "wide" {
		if len(data) == 0 {
			fmt.Fprintln(os.Stderr, "No resources found.")
			os.Exit(1)
		} else {
			if vars.ShowLabelsBoolVar {
				headers = append(headers, "labels")
			}
			helpers.PrintTable(headers, data)
		}
	} else {
		if len(returnObjects.Items) == 0 {
			fmt.Fprintln(os.Stderr, "No resources found.")
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
		} else if vars.OutputStringVar == "name" {
			for _, obj := range returnObjects.Items {
				fmt.Println(strings.ToLower(obj.GetKind()) + "/" + obj.GetName())
			}
		}
	}
	return true
}

func gatherObjects(resourcePath string, crd *apiextensionsv1.CustomResourceDefinition, objectNames []string) {
	_, err := Exists(resourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	var JSONPaths []apiextensionsv1.CustomResourceColumnDefinition
	for _, column := range crd.Spec.Versions[0].AdditionalPrinterColumns {
		JSONPaths = append(JSONPaths, column)
	}
	resources, _ := ioutil.ReadDir(resourcePath)
	for _, f := range resources {
		if f.IsDir() {
			continue
		}
		resourceYamlPath := resourcePath + f.Name()
		resourceByte, _ := ioutil.ReadFile(resourceYamlPath)
		unstruct := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(resourceByte), &unstruct); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if unstruct.IsList() {
			unstructList := &unstructured.UnstructuredList{}
			err := yaml.Unmarshal([]byte(resourceByte), &unstructList)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			for _, resource := range unstructList.Items {
				var resourceData []string
				resourceName := resource.GetName()
				if resource.GetKind() == crd.Spec.Names.Kind {
					labels := helpers.ExtractLabels(unstruct.GetLabels())
					if helpers.MatchLabels(labels, vars.LabelSelectorStringVar) && (len(objectNames) == 0 || helpers.StringInSlice(resourceName, objectNames)) {
						if vars.OutputStringVar == "" || vars.OutputStringVar == "wide" {
							if vars.AllNamespaceBoolVar && crd.Spec.Scope != "Cluster" {
								resourceData = append(resourceData, resource.GetNamespace())
							}
							resourceData = append(resourceData, resourceName)
							if crd.Spec.Versions[0].AdditionalPrinterColumns == nil {
								unstruct := &unstructured.Unstructured{}
								if err := yaml.Unmarshal([]byte(resourceByte), &unstruct); err != nil {
									fmt.Fprintln(os.Stderr, "File:", resourcePath, " does not contain a valid k8s object,", err.Error())
									os.Exit(1)
								}
								v := helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
								resourceData = append(resourceData, v)
							} else {
								for _, column := range crd.Spec.Versions[0].AdditionalPrinterColumns {
									if column.Priority == 0 || vars.OutputStringVar == "wide" {
										v := getFromJsonPath(resource.Object, "{"+column.JSONPath+"}")
										if column.Type == "date" {
											v = helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
										}
										resourceData = append(resourceData, v)
									}
								}
							}
							if vars.ShowLabelsBoolVar {
								resourceData = append(resourceData, labels)
							}
						} else {
							returnObjects.Items = append(returnObjects.Items, *unstruct)
						}
					}
				}
				data = append(data, resourceData)
			}
		} else {
			var resourceData []string
			if unstruct.GetKind() == crd.Spec.Names.Kind {
				resourceName := unstruct.GetName()
				labels := helpers.ExtractLabels(unstruct.GetLabels())
				if helpers.MatchLabels(labels, vars.LabelSelectorStringVar) && (len(objectNames) == 0 || helpers.StringInSlice(resourceName, objectNames)) {
					if vars.OutputStringVar == "" || vars.OutputStringVar == "wide" {
						if vars.AllNamespaceBoolVar && crd.Spec.Scope != "Cluster" {
							resourceData = append(resourceData, unstruct.GetNamespace())
						}
						resourceData = append(resourceData, resourceName)
						for _, jpath := range JSONPaths {
							if jpath.Priority == 0 || vars.OutputStringVar == "wide" {
								v := getFromJsonPath(unstruct.Object, "{"+jpath.JSONPath+"}")
								if jpath.Type == "date" {
									v = helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
								}
								resourceData = append(resourceData, v)
							}
						}
						if crd.Spec.Versions[0].AdditionalPrinterColumns == nil {
							age := helpers.GetAge(resourceYamlPath, unstruct.GetCreationTimestamp())
							resourceData = append(resourceData, age)
						}
						if vars.ShowLabelsBoolVar {
							resourceData = append(resourceData, labels)
						}
						data = append(data, resourceData)
					} else {
						returnObjects.Items = append(returnObjects.Items, *unstruct)
					}
				}
			}
		}
	}
}

func getFromJsonPath(data interface{}, jsonPathTemplate string) string {
	buf := new(bytes.Buffer)
	jPath := jsonpath.New("out")
	jPath.AllowMissingKeys(false)
	jPath.EnableJSONOutput(false)
	err := jPath.Parse(jsonPathTemplate)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: error parsing jsonpath "+jsonPathTemplate+", "+err.Error())
		os.Exit(1)
	}
	jPath.Execute(buf, data)
	return buf.String()
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
