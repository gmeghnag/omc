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
package get

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	goyaml "gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cliprint "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/klog/v2"
	"k8s.io/kube-aggregator/pkg/apis/apiregistration"
	"k8s.io/kubernetes/pkg/printers"
	"sigs.k8s.io/yaml"

	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	corev1 "k8s.io/api/core/v1"

	"os"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/pkg/deserializer"
	"github.com/gmeghnag/omc/pkg/tablegenerator"
	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	_ "embed"

	"github.com/spf13/cobra"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	appsv1printer "github.com/openshift/openshift-apiserver/pkg/apps/printers/internalversion"
	authorizationprinters "github.com/openshift/openshift-apiserver/pkg/authorization/printers/internalversion"
	buildprinters "github.com/openshift/openshift-apiserver/pkg/build/printers/internalversion"
	imageprinters "github.com/openshift/openshift-apiserver/pkg/image/printers/internalversion"
	projectprinters "github.com/openshift/openshift-apiserver/pkg/project/printers/internalversion"
	quotaprinters "github.com/openshift/openshift-apiserver/pkg/quota/printers/internalversion"
	routeprinters "github.com/openshift/openshift-apiserver/pkg/route/printers/internalversion"
	securityprinters "github.com/openshift/openshift-apiserver/pkg/security/printers/internalversion"
	templateprinters "github.com/openshift/openshift-apiserver/pkg/template/printers/internalversion"
)

var outputStringVar string
var allNamespaceBoolVar, showLabelsBoolVar bool
var emptyslice []string
var resourcesAndObjects [][]string

//go:embed known-resources.yaml
var yamlData []byte

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get kubernetes/openshift object in tabular format or wide|yaml|json|jsonpath.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		if vars.OutputStringVar == "wide" {
			vars.Wide = true
		}
		err := validateArgs(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for resource := range vars.GetArgs {
			resourceNamePlural, resourceGroup, namespaced, err := kindGroupNamespaced(resource)
			if err != nil {
				klog.V(1).ErrorS(err, "ERROR")
				os.Exit(1)
			}
			// namespaces, clusterloggings and clusterlogforwarders locations
			// are exceptions to must-gather resources structure
			if resourceNamePlural == "namespaces" || resourceNamePlural == "projects" {
				getNamespacesResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else if resourceNamePlural == "clusterloggings" {
				getClusterLoggingResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else if resourceNamePlural == "clusterlogforwarders" {
				getClusterLogForwarderResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else if namespaced {
				getNamespacedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else {
				getClusterScopedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			}
		}
		handleOutput()
	},
}

func init() {
	GetCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	GetCmd.PersistentFlags().BoolVar(&vars.NoHeaders, "no-headers", false, "When using the default or custom-column output format, don't print headers (default print headers).")
	GetCmd.PersistentFlags().BoolVar(&vars.ShowManagedFields, "show-managed-fields", false, "If true, show the managedFields when printing objects in JSON or YAML format.")
	GetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	GetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath=...")
	GetCmd.PersistentFlags().StringVarP(&vars.LabelSelectorStringVar, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
}

func init() {
	vars.GetArgs = make(map[string]map[string]struct{})
	vars.AliasToCrd = make(map[string]apiextensionsv1.CustomResourceDefinition)
	vars.ArgPresent = make(map[string]bool)
	vars.KnownResources = make(map[string]map[string]interface{})
	vars.UnstructuredList = types.UnstructuredList{Kind: "List", ApiVersion: "v1", Items: []unstructured.Unstructured{}}
	vars.JsonPathList = types.JsonPathList{Kind: "List", ApiVersion: "v1"}
	err := goyaml.Unmarshal(yamlData, vars.KnownResources)
	if err != nil {
		fmt.Println(err)
	}
	vars.Schema = runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		metav1beta1.AddMetaToScheme,
		corev1.AddToScheme,
		apiregistration.AddToScheme,
	}
	_ = addAdmissionRegistrationTypes(vars.Schema)
	_ = addApiServerInternalTypes(vars.Schema)
	_ = addApiRegistrationTypes(vars.Schema)
	_ = addAppsTypes(vars.Schema)
	_ = addAppsV1Types(vars.Schema)
	_ = addAuthorizationTypes(vars.Schema)
	_ = addAutoscalingTypes(vars.Schema)
	_ = addBatchTypes(vars.Schema)
	_ = addBuildTypes(vars.Schema)
	_ = addCertificatesTypes(vars.Schema)
	_ = addConfigV1Types(vars.Schema)
	_ = addCoordinationTypes(vars.Schema)
	_ = addDiscoveryTypes(vars.Schema)
	_ = addFlowControlTypes(vars.Schema)
	_ = addFlowControlV1B2Types(vars.Schema)
	_ = addImageTypes(vars.Schema)
	_ = addNetworkingTypes(vars.Schema)
	_ = addNodeTypes(vars.Schema)
	_ = addPolicyV1Types(vars.Schema)
	_ = addPolicyV1B1Types(vars.Schema)
	_ = addProjectV1Types(vars.Schema)
	_ = addQuotaV1Types(vars.Schema)
	_ = addResourceV1A2Types(vars.Schema)
	_ = addRouteV1Types(vars.Schema)
	_ = addRBACTypes(vars.Schema)
	_ = addSchedulingTypes(vars.Schema)
	_ = addSecurityV1Types(vars.Schema)
	_ = addStorageV1Types(vars.Schema)
	_ = addStorageV1B1Types(vars.Schema)
	_ = addTemplateV1Types(vars.Schema)
	utilruntime.Must(schemeBuilder.AddToScheme(vars.Schema))

	vars.TableGenerator = printers.NewTableGenerator()

	AddMissingHandlers(vars.TableGenerator)
	printersinternal.AddHandlers(vars.TableGenerator)
	buildprinters.AddBuildOpenShiftHandlers(vars.TableGenerator)
	appsv1printer.AddAppsOpenShiftHandlers(vars.TableGenerator)
	authorizationprinters.AddAuthorizationOpenShiftHandler(vars.TableGenerator)
	imageprinters.AddImageOpenShiftHandlers(vars.TableGenerator)
	projectprinters.AddProjectOpenShiftHandlers(vars.TableGenerator)
	quotaprinters.AddQuotaOpenShiftHandler(vars.TableGenerator)
	securityprinters.AddSecurityOpenShiftHandler(vars.TableGenerator)
	routeprinters.AddRouteOpenShiftHandlers(vars.TableGenerator)
	templateprinters.AddTemplateOpenShiftHandlers(vars.TableGenerator)
}

func getNamespacedResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	var namespaces []string
	if vars.AllNamespaceBoolVar == true {
		vars.Namespace = ""
		vars.ShowNamespace = true
		_namespaces, _ := ioutil.ReadDir(vars.MustGatherRootPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, vars.Namespace)
	}
	for _, namespace := range namespaces {
		UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}

		resourcePath := fmt.Sprintf("%s/namespaces/%s/%s/%s.yaml", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural)
		_file, err := ioutil.ReadFile(resourcePath)
		if err != nil {
			resourceDir := fmt.Sprintf("%s/namespaces/%s/%s/%s", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural)
			_, err = os.Stat(resourceDir)
			if err != nil {
				resourceDir = fmt.Sprintf("%s/namespaces/%s/%s/%s.%s", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural, resourceGroup)
				_, err = os.Stat(resourceDir)
			}
			if err == nil {
				resourcesFiles, _ := ioutil.ReadDir(resourceDir)
				for _, f := range resourcesFiles {
					resourceYamlPath := resourceDir + "/" + f.Name()
					_file, _ := ioutil.ReadFile(resourceYamlPath)
					item := unstructured.Unstructured{}
					if err := yaml.Unmarshal(_file, &item); err != nil {
						fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
						os.Exit(1)
					}
					if len(resources) > 0 {
						_, ok := resources[item.GetName()]
						if ok {
							handleObject(item)
						}
					} else {
						handleObject(item)
					}
				}
			} else {
				klog.V(3).Info("INFO ", fmt.Sprintf("failed to find resources for: %s", resourceNamePlural))
			}
		} else {
			if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			for _, item := range UnstructuredItems.Items {
				if len(resources) > 0 {
					_, ok := resources[item.GetName()]
					if ok {
						handleObject(item)
					}
				} else {
					handleObject(item)
				}
			}
		}
	}
}

func getClusterLogForwarderResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	resourcesYamlPath := vars.MustGatherRootPath + "/cluster-logging/clo/clusterlogforwarder_instance.yaml"
	_file, err := ioutil.ReadFile(resourcesYamlPath)
	if err == nil {
		UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourcesYamlPath)
			os.Exit(1)
		}
		for _, item := range UnstructuredItems.Items {
			_, ok := resources[item.GetName()]
			if ok || len(resources) == 0 {
				handleObject(item)
			}
		}
	}
}

func getClusterLoggingResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	resourcesYamlPath := vars.MustGatherRootPath + "/cluster-logging/clo/clusterlogging_instance.yaml"
	_file, err := ioutil.ReadFile(resourcesYamlPath)
	if err == nil {
		UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourcesYamlPath)
			os.Exit(1)
		}
		for _, item := range UnstructuredItems.Items {
			_, ok := resources[item.GetName()]
			if ok || len(resources) == 0 {
				handleObject(item)
			}
		}
	}
}

func getNamespacesResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	if len(resources) > 0 {
		for namespace := range resources {
			resourceYamlPath := fmt.Sprintf("%s/namespaces/%s/%s.yaml", vars.MustGatherRootPath, namespace, namespace)
			_file, err := ioutil.ReadFile(resourceYamlPath)
			if err == nil {
				item := unstructured.Unstructured{}
				if err := yaml.Unmarshal(_file, &item); err != nil {
					fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
					os.Exit(1)
				}
				handleObject(item)
			}
		}
	} else {
		_namespaces, _ := ioutil.ReadDir(vars.MustGatherRootPath + "/namespaces/")
		for _, namespace := range _namespaces {
			resourceYamlPath := fmt.Sprintf("%s/namespaces/%s/%s.yaml", vars.MustGatherRootPath, namespace.Name(), namespace.Name())
			_file, err := ioutil.ReadFile(resourceYamlPath)
			if err == nil {
				item := unstructured.Unstructured{}
				if err := yaml.Unmarshal(_file, &item); err != nil {
					fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
					os.Exit(1)
				}
				handleObject(item)
			}
		}
	}
}

func getClusterScopedResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
	resourcePath := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s.yaml", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
	_file, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		resourceDir := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
		resourcesFiles, _ := ioutil.ReadDir(resourceDir)
		for _, f := range resourcesFiles {
			resourceYamlPath := resourceDir + "/" + f.Name()
			_file, _ := ioutil.ReadFile(resourceYamlPath)
			item := unstructured.Unstructured{}
			if err := yaml.Unmarshal(_file, &item); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
				os.Exit(1)
			}
			if len(resources) > 0 {
				_, ok := resources[item.GetName()]
				if ok {
					handleObject(item)
				}
			} else {
				handleObject(item)
			}
		}
	} else {
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, item := range UnstructuredItems.Items {
			if len(resources) > 0 {
				_, ok := resources[item.GetName()]
				if ok {
					handleObject(item)
				}
			} else {
				handleObject(item)
			}
		}
	}

}

func handleObject(obj unstructured.Unstructured) error {
	if vars.Namespace != "" && obj.GetNamespace() != "" && vars.Namespace != obj.GetNamespace() {
		return nil
	}
	labelsOk, err := helpers.MatchLabelsFromMap(obj.GetLabels(), vars.LabelSelectorStringVar)
	if !labelsOk {
		return nil
	}
	vars.LastKind = obj.GetKind()
	if vars.OutputStringVar == "yaml" || vars.OutputStringVar == "json" {
		if vars.ShowManagedFields == false {
			obj.SetManagedFields(nil)
		}
		vars.UnstructuredList.Items = append(vars.UnstructuredList.Items, obj)
		return nil
	}
	if strings.HasPrefix(vars.OutputStringVar, "jsonpath=") {
		if vars.ShowManagedFields == false {
			obj.SetManagedFields(nil)
		}
		vars.UnstructuredList.Items = append(vars.UnstructuredList.Items, obj)
		vars.JsonPathList.Items = append(vars.JsonPathList.Items, obj.Object)
		return nil
	}
	if vars.OutputStringVar == "name" {
		if obj.GetAPIVersion() == "v1" {
			vars.Output.WriteString(strings.ToLower(obj.GetKind()) + "/" + obj.GetName() + "\n")
		} else {
			vars.Output.WriteString(strings.ToLower(obj.GetKind()) + "." + strings.Split(obj.GetAPIVersion(), "/")[0] + "/" + obj.GetName() + "\n")
		}
		return nil
	}
	rawObject, err := yaml.Marshal(obj.Object)
	if err != nil {
		klog.V(1).ErrorS(err, err.Error())
		return err
	}
	klog.V(3).Info("INFO deserializing ", obj.GetKind(), " ", obj.GetName())
	var objectTable *metav1.Table
	_, ok := vars.KnownResources[strings.ToLower(obj.GetKind())]
	if ok {
		runtimeObjectType := deserializer.RawObjectToRuntimeObject(rawObject, vars.Schema)
		if err := yaml.Unmarshal(rawObject, runtimeObjectType); err != nil {
			klog.V(3).Info(err, err.Error())
		}
		objectTable, err = tablegenerator.InternalResourceTable(runtimeObjectType, &obj)
		if err != nil {
			klog.V(3).Info("INFO ", fmt.Sprintf("%s: %s, %s", err.Error(), obj.GetKind(), obj.GetAPIVersion()))
			klog.V(1).ErrorS(err, err.Error())
			return err
		}
	} else {
		objectTable, err = tablegenerator.GenerateCustomResourceTable(obj)
		if err != nil {
			klog.V(1).ErrorS(err, err.Error())
			return err
		}
	}

	if vars.CurrentKind == obj.GetObjectKind().GroupVersionKind().Kind {
		vars.Table.Rows = append(vars.Table.Rows, objectTable.Rows...)
	} else {
		// printo la tabella dell'oggetto precedente
		printer := cliprint.NewTablePrinter(cliprint.PrintOptions{NoHeaders: vars.NoHeaders, Wide: vars.Wide, WithNamespace: false, ShowLabels: false})
		err = printer.PrintObj(&vars.Table, &vars.Output)
		if err != nil {
			klog.V(1).ErrorS(err, err.Error())
			return err
		}
		if vars.CurrentKind != "" {
			vars.Output.WriteByte('\n')
		}
		vars.CurrentKind = obj.GetObjectKind().GroupVersionKind().Kind
		vars.Table = metav1.Table{ColumnDefinitions: objectTable.ColumnDefinitions, Rows: objectTable.Rows}
	}
	return nil
}

func handleOutput() {
	printer := cliprint.NewTablePrinter(cliprint.PrintOptions{NoHeaders: vars.NoHeaders, Wide: vars.Wide, WithNamespace: false, ShowLabels: false})
	_resources := make([]string, 0, len(vars.GetArgs))
	for resource := range vars.GetArgs {
		_resources = append(_resources, resource)
	}
	resources := strings.Join(_resources, ",")
	if vars.OutputStringVar == "json" {
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			data, _ := json.MarshalIndent(vars.UnstructuredList.Items[0].Object, "", "  ")
			data = append(data, '\n')
			fmt.Printf("%s", data)
		} else if !vars.SingleResource && len(vars.UnstructuredList.Items) > 0 {
			data, _ := json.MarshalIndent(vars.UnstructuredList, "", "  ")
			data = append(data, '\n')
			fmt.Printf("%s", data)
		} else {
			if vars.Namespace != "" {
				fmt.Printf("No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Printf("No resources %s found.\n", resources)
			}
		}
	} else if strings.HasPrefix(vars.OutputStringVar, "jsonpath=") {
		jsonPathTemplate := helpers.GetJsonTemplate(vars.OutputStringVar)
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			helpers.ExecuteJsonPath(vars.UnstructuredList.Items[0].Object, jsonPathTemplate)
		} else if !vars.SingleResource && len(vars.UnstructuredList.Items) > 0 {
			helpers.ExecuteJsonPath(vars.JsonPathList, jsonPathTemplate)
		} else {
			if vars.Namespace != "" {
				fmt.Printf("No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Printf("No resources %s found.\n", resources)
			}
		}
	} else if vars.OutputStringVar == "yaml" {
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			data, _ := yaml.Marshal(vars.UnstructuredList.Items[0].Object)
			fmt.Printf("%s", data)
		} else if len(vars.UnstructuredList.Items) > 0 {
			data, _ := yaml.Marshal(vars.UnstructuredList)
			fmt.Printf("%s", data)
		} else {
			if vars.Namespace != "" {
				fmt.Printf("No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Printf("No resources %s found.\n", resources)
			}
		}
	} else {
		if vars.LastKind == vars.CurrentKind {
			err := printer.PrintObj(&vars.Table, &vars.Output)
			if err != nil {
				klog.V(1).ErrorS(err, "ERROR")
			}
			vars.Table = metav1.Table{}
		}
		if vars.Output.Len() == 0 {
			if vars.Namespace != "" {
				fmt.Printf("No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Printf("No resources %s found.\n", resources)
			}
		} else {
			vars.Output.WriteTo(os.Stdout)
		}
	}
}
