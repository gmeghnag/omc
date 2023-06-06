package get

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gmeghnag/omc/vars"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func validateArgs(args []string) error {
	var _args []string
	for _, arg := range args {
		_args = append(_args, strings.ToLower(arg))
	}
	args = _args
	if len(args) == 1 && !strings.Contains(args[0], "/") {
		if strings.Contains(args[0], ",") {
			vars.ShowKind = true
			resourcesTypes := strings.Split(strings.TrimPrefix(strings.TrimSuffix(args[0], ","), ","), ",")
			for _, resourceType := range resourcesTypes {
				if strings.Contains(resourceType, ".") {
					resourceType = strings.SplitN(resourceType, ".", 2)[0]
				}
				resourceNamePlural, _, _, err := kindGroupNamespaced(resourceType)
				if err == nil {
					vars.GetArgs[resourceNamePlural] = make(map[string]struct{})
				} else {
					return fmt.Errorf("resource type \"%s\" not known.", resourceType)
				}

			}
		} else {
			resourceType := args[0]
			if strings.Contains(args[0], ".") {
				resourceType = strings.SplitN(args[0], ".", 2)[0]
			}
			resourceNamePlural, _, _, err := kindGroupNamespaced(resourceType)
			if err == nil {
				vars.GetArgs[resourceNamePlural] = make(map[string]struct{})
			} else {
				return fmt.Errorf("resource type \"%s\" not known.", resourceType)
			}
		}
	} else if len(args) > 0 && strings.Contains(args[0], "/") {
		if len(args) == 1 {
			vars.SingleResource = true
		}
		for _, arg := range args {
			if strings.Contains(arg, "/") {
				resource := strings.Split(arg, "/")
				resourceType, resourceName := resource[0], resource[1]
				if strings.Contains(resourceType, ".") {
					resourceType = strings.SplitN(resourceType, ".", 2)[0]
				}
				resourceNamePlural, _, _, err := kindGroupNamespaced(resourceType)
				if err == nil {
					_, ok := vars.GetArgs[resourceNamePlural]
					if !ok {
						vars.GetArgs[resourceNamePlural] = make(map[string]struct{})
						vars.GetArgs[resourceNamePlural][resourceName] = struct{}{}
					} else {
						vars.GetArgs[resourceNamePlural][resourceName] = struct{}{}
					}
				} else {
					return fmt.Errorf("resource type \"%s\" not known.", resourceType)
				}
			} else {
				return fmt.Errorf("there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. 'omc get resource/<resource_name>' instead of 'omc get resource resource/<resource_name>'")
			}
		}
		if len(vars.GetArgs) > 1 {
			vars.ShowKind = true
		}
	} else if len(args) > 1 && !strings.Contains(args[0], "/") {
		resourceType := args[0]
		if strings.Contains(resourceType, ".") {
			resourceType = strings.SplitN(resourceType, ".", 1)[0]
		}
		resourceNamePlural, _, _, err := kindGroupNamespaced(resourceType)
		if err == nil {
			vars.GetArgs[resourceNamePlural] = make(map[string]struct{})
		} else {
			return fmt.Errorf("resource type \"%s\" not known.", resourceType)
		}
		if len(args[0:]) == 2 {
			vars.SingleResource = true
		}
		for _, resourceName := range args[1:] {
			if strings.Contains(resourceName, "/") {
				return fmt.Errorf("there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. 'omc get resource/<resource_name>' instead of 'omc get resource resource/<resource_name>'")
			}
			vars.GetArgs[resourceNamePlural][resourceName] = struct{}{}
		}
	}
	return nil
}

func kindGroupNamespaced(alias string) (string, string, bool, error) {
	value, ok := vars.KnownResources[alias]
	if ok {
		klog.V(3).Info("INFO ", fmt.Sprintf("Alias \"%s\" is a known resource.", alias))
		resourceNamePlural := value["plural"].(string)
		resourceGroup := value["group"].(string)
		namespaced := value["namespaced"].(bool)
		return resourceNamePlural, resourceGroup, namespaced, nil
	} else {
		klog.V(3).Info("INFO ", fmt.Sprintf("Alias \"%s\" resource not known.", alias))
		crd, ok := vars.AliasToCrd[alias]
		if ok {
			_crd := &apiextensionsv1.CustomResourceDefinition{Spec: crd.Spec}
			namespaced := false
			if _crd.Spec.Scope == "namespaced" {
				namespaced = true
			}
			return _crd.Spec.Names.Plural, _crd.Spec.Group, namespaced, nil
		}
		return kindGroupNamespacedFromCrds(alias)
	}
}

func kindGroupNamespacedFromCrds(alias string) (string, string, bool, error) {

	crdsPath := vars.MustGatherRootPath + "/cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/"
	_, err := Exists(crdsPath)
	if err == nil {
		crds, _ := ioutil.ReadDir(crdsPath)
		for _, f := range crds {
			crdYamlPath := crdsPath + f.Name()
			crdByte, _ := ioutil.ReadFile(crdYamlPath)
			_crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal([]byte(crdByte), &_crd); err != nil {
				continue
			}
			vars.AliasToCrd[strings.ToLower(_crd.Spec.Names.Kind)] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
			if strings.ToLower(_crd.Spec.Names.Kind) == alias || strings.ToLower(_crd.Spec.Names.Plural) == alias || strings.ToLower(_crd.Spec.Names.Singular) == alias || StringInSlice(alias, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == alias {
				vars.AliasToCrd[alias] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
				klog.V(4).Info("INFO ", fmt.Sprintf("Alias  \"%s\" found in path \"%s\".", alias, crdYamlPath))
				namespaced := false
				if _crd.Spec.Scope == "Namespaced" {
					namespaced = true
				}
				return _crd.Spec.Names.Plural, _crd.Spec.Group, namespaced, nil
			}
			klog.V(5).Info("INFO ", fmt.Sprintf("Alias \"%s\" not found in path \"%s\".", alias, crdYamlPath))
		}
		klog.V(4).Info("INFO ", fmt.Sprintf("No customResource found with name or alias \"%s\" in path: \"%s\".", alias, crdsPath))
	}

	home, _ := os.UserHomeDir()
	omcCrdsPath := home + "/.omc/customresourcedefinitions/"
	crds, err := ioutil.ReadDir(omcCrdsPath)
	if err != nil {
		panic(err)
	}
	for _, f := range crds {
		crdYamlPath := omcCrdsPath + f.Name()
		crdByte, _ := ioutil.ReadFile(crdYamlPath)
		_crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal([]byte(crdByte), &_crd); err != nil {
			continue
		}
		vars.AliasToCrd[strings.ToLower(_crd.Spec.Names.Kind)] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
		if strings.ToLower(_crd.Spec.Names.Kind) == alias || strings.ToLower(_crd.Spec.Names.Plural) == alias || strings.ToLower(_crd.Spec.Names.Singular) == alias || StringInSlice(alias, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == alias {
			vars.AliasToCrd[alias] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
			klog.V(4).Info("INFO ", fmt.Sprintf("Alias  \"%s\" found in path \"%s\".", alias, crdYamlPath))
			namespaced := false
			if _crd.Spec.Scope == "Namespaced" {
				namespaced = true
			}
			return _crd.Spec.Names.Plural, _crd.Spec.Group, namespaced, nil
		}
		klog.V(5).Info("INFO ", fmt.Sprintf("Alias \"%s\" not found in path \"%s\".", alias, crdYamlPath))
	}
	klog.V(4).Info("INFO ", fmt.Sprintf("No customResource found with name or alias \"%s\" in path: \"%s\".", alias, omcCrdsPath))
	return alias, "", false, fmt.Errorf("No customResource found with name or alias \"%s\".", alias)
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
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
