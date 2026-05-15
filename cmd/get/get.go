/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>
Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.

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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strings"

	goyaml "gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cliprint "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/util/jsonpath"
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

var jsonRegexp = regexp.MustCompile(`^\{\.?([^{}]+)\}$|^\.?([^{}]+)$`)

//go:embed known-resources.yaml
var yamlData []byte

var GetCmd = &cobra.Command{
	Use:          "get",
	Short:        "Get kubernetes/openshift object in tabular format or wide|yaml|json|jsonpath|custom-columns.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		if vars.OutputStringVar == "wide" {
			vars.Wide = true
		}
		if err := validateArgs(args); err != nil {
			return err
		}
		for resource := range vars.GetArgs {
			resourceNamePlural, resourceGroup, _, namespaced, err := KindGroupNamespaced(resource)
			if err != nil {
				klog.V(1).ErrorS(err, "ERROR")
				return err
			}
			// namespaces and projects resources
			// are exceptions to must-gather resources structure
			switch {
			case resourceNamePlural == "namespaces" || resourceNamePlural == "projects":
				err = getNamespacesResources(vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			case resourceNamePlural == "podnetworkconnectivitychecks":
				err = getPodNetworkConnectivityChecksResources(vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			case namespaced:
				err = getNamespacedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			default:
				err = getClusterScopedResources(resourceNamePlural, resourceGroup, vars.GetArgs[resourceNamePlural+"."+resourceGroup])
			}
			if err != nil {
				return err
			}
		}
		return handleOutput(os.Stdout, os.Stderr)
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
	_ = addAuthenticationTypes(vars.Schema)
	_ = addAuthorizationTypes(vars.Schema)
	_ = addAutoscalingV1Types(vars.Schema)
	_ = addAutoscalingV2Types(vars.Schema)
	_ = addBatchTypes(vars.Schema)
	_ = addBuildTypes(vars.Schema)
	_ = addCertificatesTypes(vars.Schema)
	_ = addConfigV1Types(vars.Schema)
	_ = addCoordinationTypes(vars.Schema)
	_ = addDiscoveryTypes(vars.Schema)
	_ = addEventsV1Types(vars.Schema)
	_ = addFlowControlTypes(vars.Schema)
	_ = addFlowControlV1B2Types(vars.Schema)
	_ = addFlowControlV1Types(vars.Schema)
	_ = addImageTypes(vars.Schema)
	_ = addNetworkingTypes(vars.Schema)
	_ = addNodeTypes(vars.Schema)
	_ = addPolicyV1Types(vars.Schema)
	_ = addPolicyV1B1Types(vars.Schema)
	_ = addProjectV1Types(vars.Schema)
	_ = addQuotaV1Types(vars.Schema)
	_ = addResourceV1A2Types(vars.Schema)
	_ = addResourceV1Types(vars.Schema)
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

func getNamespacedResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) error {
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
							return fmt.Errorf("error reading %s: %w", podPath, err)
						}
						var podItem unstructured.Unstructured
						if err := yaml.Unmarshal(_file, &podItem); err != nil {
							return fmt.Errorf("error unmarshaling %s: %w", podPath, err)
						}
						if podItem.Object != nil {
							UnstructuredItems.Items = append(UnstructuredItems.Items, podItem)
						}
					}
				} else {
					return fmt.Errorf("error unmarshaling %s: %w", resourcesItemsPath, err)
				}
			}

			if vars.SortBy != "" {
				UnstructuredItems.Items = sortResources(UnstructuredItems.Items, vars.SortBy)
			}
			for _, item := range UnstructuredItems.Items {
				if len(resources) > 0 {
					if _, ok := resources[item.GetName()]; ok {
						if err := handleObject(item); err != nil {
							return err
						}
					}
				} else {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			}
		} else { // the resources are customresources so, stored in a single file per resource
			resourceDir := fmt.Sprintf("%s/namespaces/%s/%s/%s", vars.MustGatherRootPath, namespace, resourceGroup, resourceNamePlural)
			_, err = os.Stat(resourceDir)
			if err == nil {
				resourcesFiles, rErr := ReadDirForResources(resourceDir)
				if rErr != nil {
					klog.V(3).ErrorS(err, "Failed to read resources:")
				}
				var sortObjects []unstructured.Unstructured
				for _, f := range resourcesFiles {
					if f.IsDir() {
						fmt.Fprintf(
							os.Stderr,
							"error: invalid must-gather structure, yaml files are expected in path \"/namespaces/%s/%s/%s\", found directory: \"%s\"\n",
							namespace, resourceGroup, resourceNamePlural, f.Name(),
						)
						continue
					}
					resourceYamlPath := resourceDir + "/" + f.Name()
					_file, err := os.ReadFile(resourceYamlPath)
					if err != nil {
						return fmt.Errorf("error reading %s: %w", resourceYamlPath, err)
					}
					item := unstructured.Unstructured{}
					if err := yaml.Unmarshal(_file, &item); err != nil {
						return fmt.Errorf("error unmarshaling %s: %w", resourceYamlPath, err)
					}
					if vars.SortBy != "" {
						sortObjects = append(sortObjects, item)
					} else {
						if len(resources) > 0 {
							if _, ok := resources[item.GetName()]; ok {
								if err := handleObject(item); err != nil {
									return err
								}
							}
						} else {
							if err := handleObject(item); err != nil {
								return err
							}
						}
					}
				}

				if vars.SortBy != "" {
					sortObjects = sortResources(sortObjects, vars.SortBy)
					for _, item := range sortObjects {
						if len(resources) > 0 {
							if _, ok := resources[item.GetName()]; ok {
								if err := handleObject(item); err != nil {
									return err
								}
							}
						} else {
							if err := handleObject(item); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func getNamespacesResources(resources map[string]struct{}) error {
	var sortObjects []unstructured.Unstructured
	if len(resources) > 0 {
		for namespace := range resources {
			resourceYamlPath := fmt.Sprintf("%s/namespaces/%s/%s.yaml", vars.MustGatherRootPath, namespace, namespace)
			_file, err := os.ReadFile(resourceYamlPath)
			if err == nil {
				item := unstructured.Unstructured{}
				if err := yaml.Unmarshal(_file, &item); err != nil {
					return fmt.Errorf("error unmarshaling %s: %w", resourceYamlPath, err)
				}
				if vars.SortBy != "" {
					sortObjects = append(sortObjects, item)
				} else {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			}
		}
	} else {
		_namespaces, _ := os.ReadDir(vars.MustGatherRootPath + "/namespaces/")
		for _, namespace := range _namespaces {
			resourceYamlPath := fmt.Sprintf("%s/namespaces/%s/%s.yaml", vars.MustGatherRootPath, namespace.Name(), namespace.Name())
			_file, err := os.ReadFile(resourceYamlPath)
			if err == nil {
				item := unstructured.Unstructured{}
				if err := yaml.Unmarshal(_file, &item); err != nil {
					return fmt.Errorf("error unmarshaling %s: %w", resourceYamlPath, err)
				}
				if vars.SortBy != "" {
					sortObjects = append(sortObjects, item)
				} else {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			}
		}
	}
	if vars.SortBy != "" {
		sortObjects = sortResources(sortObjects, vars.SortBy)
		for _, item := range sortObjects {
			if err := handleObject(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func getClusterScopedResources(resourceNamePlural string, resourceGroup string, resources map[string]struct{}) error {
	UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
	resourcePath := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s.yaml", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
	_file, err := os.ReadFile(resourcePath)
	if err != nil {
		resourceDir := fmt.Sprintf("%s/cluster-scoped-resources/%s/%s", vars.MustGatherRootPath, resourceGroup, resourceNamePlural)
		resourcesFiles, rErr := ReadDirForResources(resourceDir)
		if rErr != nil {
			klog.V(3).ErrorS(err, "Failed to read resources:")
		}
		for _, f := range resourcesFiles {
			resourceYamlPath := resourceDir + "/" + f.Name()
			_file, err := os.ReadFile(resourceYamlPath)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", resourceYamlPath, err)
			}
			item := unstructured.Unstructured{}
			if err := yaml.Unmarshal(_file, &item); err != nil {
				return fmt.Errorf("error unmarshaling %s: %w", resourceYamlPath, err)
			}
			if item.IsList() {
				return fmt.Errorf("file %q contains a \"List\" objectKind, while it should contain a single resource", resourceYamlPath)
			}
			if vars.SortBy != "" {
				UnstructuredItems.Items = append(UnstructuredItems.Items, item)
			} else {
				if len(resources) > 0 {
					if _, ok := resources[item.GetName()]; ok {
						if err := handleObject(item); err != nil {
							return err
						}
					}
				} else {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			}
		}
		if vars.SortBy != "" {
			UnstructuredItems.Items = sortResources(UnstructuredItems.Items, vars.SortBy)
			for _, item := range UnstructuredItems.Items {
				if len(resources) > 0 {
					if _, ok := resources[item.GetName()]; ok {
						if err := handleObject(item); err != nil {
							return err
						}
					}
				} else {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			}
		}
	} else {
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			return fmt.Errorf("error unmarshaling %s: %w", resourcePath, err)
		}
		if vars.SortBy != "" {
			UnstructuredItems.Items = sortResources(UnstructuredItems.Items, vars.SortBy)
		}
		for _, item := range UnstructuredItems.Items {
			if len(resources) > 0 {
				if _, ok := resources[item.GetName()]; ok {
					if err := handleObject(item); err != nil {
						return err
					}
				}
			} else {
				if err := handleObject(item); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleObject(obj unstructured.Unstructured) error {
	if vars.Namespace != "" && obj.GetNamespace() != "" && vars.Namespace != obj.GetNamespace() {
		return nil
	}
	labelsOk, err := helpers.MatchLabelsFromMap(obj.GetLabels(), vars.LabelSelectorStringVar)
	if err != nil {
		return fmt.Errorf("invalid label selector %q: %w", vars.LabelSelectorStringVar, err)
	}
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

func handleOutput(w io.Writer, errOut io.Writer) error {
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
				fmt.Fprintf(errOut, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(errOut, "No resources %s found.\n", resources)
			}
		}
	} else if strings.HasPrefix(vars.OutputStringVar, "jsonpath=") {
		jsonPathTemplate, err := helpers.GetJsonTemplate(vars.OutputStringVar)
		if err != nil {
			return err
		}
		if vars.SingleResource && len(vars.UnstructuredList.Items) == 1 {
			if err := helpers.ExecuteJsonPath(vars.UnstructuredList.Items[0].Object, jsonPathTemplate); err != nil {
				return err
			}
		} else if !vars.SingleResource && len(vars.UnstructuredList.Items) > 0 {
			if err := helpers.ExecuteJsonPath(vars.JsonPathList, jsonPathTemplate); err != nil {
				return err
			}
		} else {
			if vars.Namespace != "" {
				fmt.Fprintf(errOut, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(errOut, "No resources %s found.\n", resources)
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
				fmt.Fprintf(errOut, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			} else {
				fmt.Fprintf(errOut, "No resources %s found.\n", resources)
			}
		}
	} else {
		if vars.LastKind == vars.CurrentKind {
			if err := printer.PrintObj(&vars.Table, &vars.Output); err != nil {
				return fmt.Errorf("error printing table: %w", err)
			}
			vars.Table = metav1.Table{}
		}
		if vars.Output.Len() == 0 {
			// never print the (default/current) namespace if at least one cluster-scoped resource is requested
			if vars.Namespace == "" || includesClusterScoped {
				fmt.Fprintf(errOut, "No resources %s found.\n", resources)
			} else {
				fmt.Fprintf(errOut, "No resources %s found in %s namespace.\n", resources, vars.Namespace)
			}
		} else {
			vars.Output.WriteTo(w)
		}
	}
	return nil
}

func getPodNetworkConnectivityChecksResources(resources map[string]struct{}) error {
	resourcesYamlPath := vars.MustGatherRootPath + "/pod_network_connectivity_check/podnetworkconnectivitychecks.yaml"
	_file, err := os.ReadFile(resourcesYamlPath)
	if err == nil {
		UnstructuredItems := types.UnstructuredList{ApiVersion: "v1", Kind: "List"}
		if err := yaml.Unmarshal(_file, &UnstructuredItems); err != nil {
			return fmt.Errorf("error unmarshaling %s: %w", resourcesYamlPath, err)
		}
		for _, item := range UnstructuredItems.Items {
			_, ok := resources[item.GetName()]
			if ok || len(resources) == 0 {
				if err := handleObject(item); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func sortResources(list []unstructured.Unstructured, sortBy string) []unstructured.Unstructured {
	if sortBy == "" {
		// no need to sort. Return original list
		return list
	}

	// relaxed jsonpath like kubectl/oc
	submatches := jsonRegexp.FindStringSubmatch(sortBy)
	if submatches == nil {
		klog.V(3).Info("Failed to identify relaxed jsonpath, skipping")
		return list
	}
	if len(submatches) != 3 {
		klog.V(3).Info("unexpected submatch list: ", submatches)
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
	newlist = append(newlist, list...)

	slices.SortFunc(newlist, func(a, b unstructured.Unstructured) int {
		abuf := new(bytes.Buffer)
		bbuf := new(bytes.Buffer)

		err := jpath.Execute(abuf, a.UnstructuredContent())
		if err != nil {
			klog.V(3).ErrorS(fmt.Errorf("field %s not found in object %s/%s", vars.SortBy, b.GetKind(), b.GetName()), "jsonpath error")
			return 0
		}
		err = jpath.Execute(bbuf, b.UnstructuredContent())
		if err != nil {
			klog.V(3).ErrorS(fmt.Errorf("field %s not found in object %s/%s", vars.SortBy, b.GetKind(), b.GetName()), "jsonpath error")
			return 0
		}

		return strings.Compare(abuf.String(), bbuf.String())
	})

	return newlist
}
