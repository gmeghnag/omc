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
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"slices"
	"bytes"
	"regexp"

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
	"k8s.io/client-go/util/jsonpath"
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

var jsonRegexp = regexp.MustCompile(`^\{\.?([^{}]+)\}$|^\.?([^{}]+)$`)

//go:embed known-resources.yaml
var yamlData []byte

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get kubernetes/openshift object in tabular format or wide|yaml|json|jsonpath|custom-columns.",
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
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		for resource := range vars.GetArgs {
			resourceNamePlural, resourceGroup, _, namespaced, err := KindGroupNamespaced(resource)
			if err != nil {
				klog.V(1).ErrorS(err, "ERROR")
				os.Exit(1)
			}
			// namespaces and projects resources
			// are exceptions to must-gather resources structure
			if resourceNamePlural == "namespaces" || resourceNamePlural == "projects" {
				getNamespacesResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else if resourceNamePlural == "podnetworkconnectivitychecks" {
				getPodNetworkConnectivityChecksResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else if namespaced {
				getNamespacedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			} else {
				getClusterScopedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			}
		}
		handleOutput(os.Stdout)
	},
}

func init() {
	GetCmd.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces.")
	GetCmd.PersistentFlags().BoolVar(&vars.NoHeaders, "no-headers", false, "When using the default or custom-column output format, don't print headers (default print headers).")
	GetCmd.PersistentFlags().BoolVar(&vars.ShowManagedFields, "show-managed-fields", false, "If true, show the managedFields when printing objects in JSON or YAML format.")
	GetCmd.PersistentFlags().BoolVarP(&vars.ShowLabelsBoolVar, "show-labels", "", false, "When printing, show all labels as the last column (default hide labels column)")
	GetCmd.PersistentFlags().StringVarP(&vars.OutputStringVar, "output", "o", "", "Output format. One of: json|yaml|wide|jsonpath|custom-columns=...")
	GetCmd.PersistentFlags().StringVarP(&vars.LabelSelectorStringVar, "selector", "l", "", "selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	GetCmd.PersistentFlags().StringVarP(&vars.SortBy, "sort-by", "", "", "If non-empty, sort list types using this field specification. The field specification is expressed as a JSONPath expression (e.g. '{.metadata.name}').")
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
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
	vars.Schema = runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		metav1beta1.AddMetaToScheme,
		corev1.AddToScheme,
		apiregistration.AddToScheme,
	}
	_ = addAdmissionRegistrationTypes(vars.Schema)
	_ = addApiextensionsTypes(vars.Schema)
	_ = addApiextensionsV1Beta1Types(vars.Schema)
	_ = addApiServerInternalTypes(vars.Schema)
	_ = addApiRegistrationTypes(vars.Schema)
	_ = addAppsTypes(vars.Schema)
	_ = addAppsV1Types(vars.Schema)
	_ = addAuthorizationTypes(vars.Schema)
	_ = addAutoscalingV1Types(vars.Schema)
	_ = addAutoscalingV2Types(vars.Schema)
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
	_ = addOAuthV1Types(vars.Schema)
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
	var objects []unstructured.Unstructured
	var namespaces []string
	if vars.AllNamespaceBoolVar {
		vars.Namespace = ""
		vars.ShowNamespace = true
		_namespaces, _ := ReadDirForResources(vars.MustGatherRootPath + "/namespaces/")
		for _, f := range _namespaces {
			namespaces = append(namespaces, f.Name())
		}
	} else {
		namespaces = append(namespaces, vars.Namespace)
	}
	for _, namespace := range namespaces {
		UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
		resourcesItemsPath := fmt.Sprintf("%s/namespaces/%s/%s/%s.yaml", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural)
		_file, err := os.ReadFile(resourcesItemsPath)
		if err == nil { // able to read <resourceplural>.yaml, which contains list of items, i.e. /namespaces/<NAMESPACE>/core/pods.yaml
			err := yaml.Unmarshal(_file, &UnstructuredItems)
			if err != nil { // unable to unmarshal the file, it may be empty or corrupted
				// We handle this situation by looking for the pod in the pods directory
				fStat, _ := os.Stat(resourcesItemsPath)
				fSize := fStat.Size()
				if resourceNamePlural == "pods" && fSize == 0 {
					// tranverse the pods directory and fill in UnstructuredItems.Items
					podsDir := fmt.Sprintf("%s/namespaces/%s/pods", vars.MustGatherRootPath, namespace)
					pods, rErr := ReadDirForResources(podsDir)
					if rErr != nil {
						klog.V(3).ErrorS(err, "Failed to read resources:")
					}
					for _, pod := range pods {
						podName := pod.Name()
						podPath := fmt.Sprintf("%s/%s/%s.yaml", podsDir, podName, podName)
						_file, err := os.ReadFile(podPath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error reading %s: %s\n", podPath, err)
							os.Exit(1)
						}
						var podItem unstructured.Unstructured
						if err := yaml.Unmarshal(_file, &podItem); err != nil {
							fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file "+podPath)
							os.Exit(1)
						}
						if podItem.Object != nil {
							UnstructuredItems.Items = append(UnstructuredItems.Items, podItem)
						}
					}
				} else {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}

			objects = append(objects, UnstructuredItems.Items...)
		} else { // the resources are customresources so, stored in a single file per resource
			resourceDir := fmt.Sprintf("%s/namespaces/%s/%s/%s", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural)
			_, err = os.Stat(resourceDir)
			if err == nil {
				var objects []unstructured.Unstructured
				resourcesFiles, rErr := ReadDirForResources(resourceDir)
				if rErr != nil {
					klog.V(3).ErrorS(err, "Failed to read resources:")
				}
				for _, f := range resourcesFiles {
					resourceYamlPath := resourceDir + "/" + f.Name()
					_file, _ := os.ReadFile(resourceYamlPath)
					item := unstructured.Unstructured{}
					if err := yaml.Unmarshal(_file, &item); err != nil {
						fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
						os.Exit(1)
					}

					objects = append(objects, item)
				}
			}

		}
	}

	objects = sortResources(objects, vars.SortBy)

	for _, item := range objects {
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

func getNamespacesResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	var objects []unstructured.Unstructured

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

				objects = append(objects, item)
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
				objects = append(objects, item)
			}
		}
	}

	objects = sortResources(objects, vars.SortBy)
	for _, item := range objects {
		handleObject(item)
	}
}

func getClusterScopedResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
	resourcePath := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s.yaml", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
	var objects []unstructured.Unstructured

	_file, err := os.ReadFile(resourcePath)
	if err != nil {
		resourceDir := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
		resourcesFiles, rErr := ReadDirForResources(resourceDir)
		if rErr != nil {
			klog.V(3).ErrorS(err, "Failed to read resources:")
		}

		for _, f := range resourcesFiles {
			resourceYamlPath := resourceDir + "/" + f.Name()

			_file, _ := os.ReadFile(resourceYamlPath)

			item := unstructured.Unstructured{}
			if err := yaml.Unmarshal(_file, &item); err != nil {
				fmt.Fprintln(os.Stderr, "Error when trying to unmarshal file: "+resourceYamlPath)
				os.Exit(1)
			}

			if item.IsList() {
				fmt.Fprintln(os.Stderr, "error: file \""+resourceYamlPath+"\" contains a \"List\" objectKind, while it should contain a single resource.")
				os.Exit(1)
			}

			objects = append(objects, item)
		}
	} else {
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		objects = append(objects, UnstructuredItems.Items...)
	}

	objects = sortResources(objects, vars.SortBy)

	for _, v := range objects {
		if len(resources) > 0 {
			_, ok := resources[v.GetName()]
			if ok {
				handleObject(v)
			}
		} else {
			handleObject(v)
		}
	}
}

func handleObject(obj unstructured.Unstructured) error {
	if vars.Namespace != "" && obj.GetNamespace() != "" && vars.Namespace != obj.GetNamespace() {
		return nil
	}
	labelsOk, _ := helpers.MatchLabelsFromMap(obj.GetLabels(), vars.LabelSelectorStringVar)
	if !labelsOk {
		return nil
	}
	vars.LastKind = obj.GetKind()
	if vars.OutputStringVar == "yaml" || vars.OutputStringVar == "json" {
		if !vars.ShowManagedFields {
			obj.SetManagedFields(nil)
		}
		vars.UnstructuredList.Items = append(vars.UnstructuredList.Items, obj)
		return nil
	}
	if strings.HasPrefix(vars.OutputStringVar, "jsonpath=") {
		if !vars.ShowManagedFields {
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
	if strings.HasPrefix(vars.OutputStringVar, "custom-columns=") {
		objectTable, err = tablegenerator.CustomColumnsTable(&obj)
		if err != nil {
			klog.V(1).ErrorS(err, err.Error())
			return err
		}
	} else {
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

func handleOutput(w io.Writer) {
	printer := cliprint.NewTablePrinter(cliprint.PrintOptions{NoHeaders: vars.NoHeaders, Wide: vars.Wide, WithNamespace: false, ShowLabels: false})
	_resources := make([]string, 0, len(vars.GetArgs))
	var includesClusterScoped bool
	for resource := range vars.GetArgs {
		_resources = append(_resources, resource)
		// if at least one resource is cluster-scoped, never include a namespace in the output if no resources are found of the kind
		_, _, _, namespaced, _ := KindGroupNamespaced(resource)
		if !namespaced {
			includesClusterScoped = true
		}
	}
	sort.Strings(_resources)
	resources := strings.Join(_resources, ",")
	if vars.OutputStringVar == "json" {
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			data, _ := json.MarshalIndent(vars.UnstructuredList.Items[0].Object, "", "  ")
			data = append(data, '\n')
			fmt.Fprintf(w, "%s", data)
		} else if !vars.SingleResource && len(vars.UnstructuredList.Items) > 0 {
			data, _ := json.MarshalIndent(vars.UnstructuredList, "", "  ")
			data = append(data, '\n')
			fmt.Fprintf(w, "%s", data)
		} else {
			if vars.Namespace != "" {
				fmt.Fprintf(w, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(w, "No resources %s found.\n", resources)
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
				fmt.Fprintf(w, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(w, "No resources %s found.\n", resources)
			}
		}
	} else if vars.OutputStringVar == "yaml" {
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			data, _ := yaml.Marshal(vars.UnstructuredList.Items[0].Object)
			fmt.Fprintf(w, "%s", data)
		} else if len(vars.UnstructuredList.Items) > 0 {
			data, _ := yaml.Marshal(vars.UnstructuredList)
			fmt.Fprintf(w, "%s", data)
		} else {
			if vars.Namespace != "" {
				fmt.Fprintf(w, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(w, "No resources %s found.\n", resources)
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
			// never print the (default/current) namespace if at least one cluster-scoped resource is requested
			if vars.Namespace == "" || includesClusterScoped {
				fmt.Fprintf(w, "No resources %s found.\n", resources)
			} else {
				fmt.Fprintf(w, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			}
		} else {
			vars.Output.WriteTo(w)
		}
	}
}

func sortResources(list []unstructured.Unstructured, sortBy string) []unstructured.Unstructured {
	if sortBy == "" {
		// no need to sort. Return original list
		return list
	}

	// relaxed jsonpath like kubectl/oc
	submatches := jsonRegexp.FindStringSubmatch(sortBy)
	if submatches == nil {
		fmt.Println("Failed to identify relaxed jsonpath, skipping")
		return list
	}
	if len(submatches) != 3 {
		fmt.Println("unexpected submatch list: ", submatches)
		return list
	}
	if len(submatches[1]) != 0 {
		sortBy = submatches[1]
	} else {
		sortBy = submatches[2]
	}
	sortBy = fmt.Sprintf("{.%s}", sortBy)

	jpath := jsonpath.New("out")
	jpath.AllowMissingKeys(false)
	jpath.EnableJSONOutput(true)
	jpath.Parse(sortBy)

	var newlist []unstructured.Unstructured
	newlist = append([]unstructured.Unstructured(nil), list...)

	fmt.Println("Running sort")
	slices.SortFunc(newlist, func(a, b unstructured.Unstructured) int {
		abuf := new(bytes.Buffer)
		bbuf := new(bytes.Buffer)

		err := jpath.Execute(abuf, a.UnstructuredContent())
		if err != nil {
			fmt.Println("error in jsonpath: ", err)
			fmt.Println("  for object:", a)
			return 0
		}
		err = jpath.Execute(bbuf, b.UnstructuredContent())
		if err != nil {
			fmt.Println("error in jsonpath: ", err)
			fmt.Println("  for object:", b)
			return 0
		}

		return strings.Compare(abuf.String(), bbuf.String())
	})

	return newlist
}

func getPodNetworkConnectivityChecksResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) {
	resourcesYamlPath := vars.MustGatherRootPath + "/pod_network_connectivity_check/podnetworkconnectivitychecks.yaml"
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
