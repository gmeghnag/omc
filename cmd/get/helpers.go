package get

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gmeghnag/omc/vars"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// crdCache holds the CRDs parsed off disk, keyed by must-gather root, shared
// across concurrent Run calls under the mutex.
var crdCache = struct {
	sync.Mutex
	byRoot map[string][]apiextensionsv1.CustomResourceDefinition
}{byRoot: make(map[string][]apiextensionsv1.CustomResourceDefinition)}

func validateArgs(opts *Options, args []string) error {
	// Local map: validateArgs only resolves aliases to populate GetArgs.
	aliasCache := make(map[string]apiextensionsv1.CustomResourceDefinition)
	if len(args) == 1 && args[0] == "all" {
		args = []string{"pods.core,services.core,daemonsets.apps,deployments.apps,replicasets.apps,statefulsets.apps,replicationcontrollers.core,deploymentconfigs.apps.openshift.io,builds.build.openshift.io,buildconfigs.build.openshift.io,jobs.batch,cronjobs.batch,routes.route.openshift.io,ingresses.networking.k8s.io,"}
	}
	var _args []string
	for _, arg := range args {
		_args = append(_args, strings.ToLower(arg))
	}
	args = _args
	if len(args) == 1 && !strings.Contains(args[0], "/") {
		if strings.Contains(args[0], ",") {
			opts.ShowKind = true
			resourcesTypes := strings.Split(strings.TrimPrefix(strings.TrimSuffix(args[0], ","), ","), ",")
			for _, resourceType := range resourcesTypes {
				resourceNamePlural, resourceGroup, _, _, err := kindGroupNamespaced(resourceType, opts.RootPath, aliasCache)
				if err == nil {
					if !strings.Contains(resourceType, ".") {
						opts.GetArgs[resourceNamePlural+"."+resourceGroup] = make(map[string]struct{})
					} else {
						opts.GetArgs[resourceType] = make(map[string]struct{})
					}
				} else {
					return fmt.Errorf("resource type \"%s\" not known.", resourceType)
				}

			}
		} else {
			resourceType := args[0]
			resourceNamePlural, resourceGroup, _, _, err := kindGroupNamespaced(resourceType, opts.RootPath, aliasCache)
			if err == nil {
				if !strings.Contains(resourceType, ".") {
					opts.GetArgs[resourceNamePlural+"."+resourceGroup] = make(map[string]struct{})
				} else {
					opts.GetArgs[resourceType] = make(map[string]struct{})
				}
			} else {
				return fmt.Errorf("resource type \"%s\" not known.", resourceType)
			}
		}
	} else if len(args) > 0 && strings.Contains(args[0], "/") {
		if len(args) == 1 {
			opts.SingleResource = true
		}
		for _, arg := range args {
			if strings.Contains(arg, "/") {
				resource := strings.Split(arg, "/")
				resourceType, resourceName := resource[0], resource[1]
				resourceNamePlural, resourceGroup, _, _, err := kindGroupNamespaced(resourceType, opts.RootPath, aliasCache)
				if err == nil {
					_, ok := opts.GetArgs[resourceNamePlural+"."+resourceGroup]
					if !ok {
						opts.GetArgs[resourceNamePlural+"."+resourceGroup] = make(map[string]struct{})
						opts.GetArgs[resourceNamePlural+"."+resourceGroup][resourceName] = struct{}{}
					} else {
						opts.GetArgs[resourceNamePlural+"."+resourceGroup][resourceName] = struct{}{}
					}
				} else {
					return fmt.Errorf("resource type \"%s\" not known.", resourceType)
				}
			} else {
				return fmt.Errorf("there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. 'omc get resource/<resource_name>' instead of 'omc get resource resource/<resource_name>'")
			}
		}
		if len(opts.GetArgs) > 1 {
			opts.ShowKind = true
		}
	} else if len(args) > 1 && !strings.Contains(args[0], "/") {
		resourceType := args[0]
		resourceNamePlural, resourceGroup, _, _, err := kindGroupNamespaced(resourceType, opts.RootPath, aliasCache)
		if err == nil {
			opts.GetArgs[resourceNamePlural+"."+resourceGroup] = make(map[string]struct{})
		} else {
			return fmt.Errorf("resource type \"%s\" not known.", resourceType)
		}
		if len(args[0:]) == 2 {
			opts.SingleResource = true
		}
		for _, resourceName := range args[1:] {
			if strings.Contains(resourceName, "/") {
				return fmt.Errorf("there is no need to specify a resource type as a separate argument when passing arguments in resource/name form (e.g. 'omc get resource/<resource_name>' instead of 'omc get resource resource/<resource_name>'")
			}
			opts.GetArgs[resourceNamePlural+"."+resourceGroup][resourceName] = struct{}{}
		}
	}
	return nil
}

// KindGroupNamespaced resolves an alias against the known resources, then the
// CRDs under rootPath. The get path uses kindGroupNamespaced with a per-Run map
// so concurrent Run callers against different roots stay isolated.
func KindGroupNamespaced(alias, rootPath string) (string, string, string, bool, error) {
	return kindGroupNamespaced(alias, rootPath, nil)
}

func kindGroupNamespaced(alias, rootPath string, aliasCache map[string]apiextensionsv1.CustomResourceDefinition) (string, string, string, bool, error) {
	// when it si called the second time
	if strings.Contains(alias, ".") {
		split := strings.Split(alias, ".")
		if len(split) > 1 {
			resourcePlural := strings.Join(split[:1], ".")
			group := strings.Join(split[1:], ".")
			value, ok := vars.KnownResources[resourcePlural]
			if ok {
				klog.V(3).Info("INFO ", fmt.Sprintf("Alias \"%s\" is a known resource.", alias))
				resourceNamePlural := value["plural"].(string)
				resourceNameSingular := value["name"].(string)
				resourceGroup := value["group"].(string)
				namespaced := value["namespaced"].(bool)
				if strings.HasPrefix(resourceGroup, group) {
					return resourceNamePlural, resourceGroup, resourceNameSingular, namespaced, nil
				}
			}
		}
	}
	value, ok := vars.KnownResources[alias]
	if ok {
		klog.V(3).Info("INFO ", fmt.Sprintf("Alias \"%s\" is a known resource.", alias))
		resourceNamePlural := value["plural"].(string)
		resourceNameSingular := value["name"].(string)
		resourceGroup := value["group"].(string)
		namespaced := value["namespaced"].(bool)
		return resourceNamePlural, resourceGroup, resourceNameSingular, namespaced, nil
	} else {
		klog.V(3).Info("INFO ", fmt.Sprintf("Alias \"%s\" resource not known.", alias))
		return kindGroupNamespacedFromCrds(alias, rootPath, aliasCache)
	}
}

func kindGroupNamespacedFromCrds(alias, rootPath string, aliasCache map[string]apiextensionsv1.CustomResourceDefinition) (string, string, string, bool, error) {
	crdCache.Lock()
	defer crdCache.Unlock()
	// A nil map (a caller that keeps no cache) gets a throwaway so the writes
	// below do not panic.
	if aliasCache == nil {
		aliasCache = make(map[string]apiextensionsv1.CustomResourceDefinition)
	}
	// Fast path: alias already resolved earlier in this Run.
	if crd, ok := aliasCache[alias]; ok {
		namespaced := crd.Spec.Scope == "Namespaced"
		return crd.Spec.Names.Plural, crd.Spec.Group, crd.Spec.Names.Singular, namespaced, nil
	}
	crdsPath := rootPath + "/cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/"
	bundleCRDs, hit := crdCache.byRoot[rootPath]
	if !hit {
		var cacheReady bool
		if ok, _ := Exists(crdsPath); ok {
			files, rErr := ReadDirForResources(crdsPath)
			if rErr != nil {
				fmt.Fprintln(os.Stderr, rErr)
			} else {
				cacheReady = true
				for _, f := range files {
					crdByte, err := ioutil.ReadFile(crdsPath + f.Name())
					if err != nil {
						cacheReady = false
						continue
					}
					_crd := &apiextensionsv1.CustomResourceDefinition{}
					if err := yaml.Unmarshal(crdByte, _crd); err != nil {
						continue
					}
					bundleCRDs = append(bundleCRDs, *_crd)
				}
			}
		} else {
			cacheReady = true
		}
		if cacheReady {
			crdCache.byRoot[rootPath] = bundleCRDs
		}
	}
	for _, _crd := range bundleCRDs {
		if strings.Contains(alias, ".") {
			split := strings.Split(alias, ".")
			if len(split) > 1 {
				group := strings.Join(split[1:], ".")
				if !strings.HasPrefix(_crd.Spec.Group, group) {
					continue
				} else {
					_alias := strings.Join(split[:1], ".")
					if strings.ToLower(_crd.Spec.Names.Plural) == _alias || strings.ToLower(_crd.Spec.Names.Singular) == _alias || StringInSlice(_alias, _crd.Spec.Names.ShortNames) {
						namespaced := false
						if _crd.Spec.Scope == "Namespaced" {
							namespaced = true
						}
						aliasCache[strings.ToLower(_crd.Spec.Names.Kind)+"."+_crd.Spec.Group] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
						return _crd.Spec.Names.Plural, _crd.Spec.Group, _crd.Spec.Names.Singular, namespaced, nil
					}
				}
			}
		}
		aliasCache[strings.ToLower(_crd.Spec.Names.Kind)+"."+_crd.Spec.Group] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
		if strings.ToLower(_crd.Spec.Names.Kind) == alias || strings.ToLower(_crd.Spec.Names.Plural) == alias || strings.ToLower(_crd.Spec.Names.Singular) == alias || StringInSlice(alias, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == alias {
			aliasCache[alias] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
			klog.V(4).Info("INFO ", fmt.Sprintf("Alias  \"%s\" found in bundle CRDs.", alias))
			namespaced := false
			if _crd.Spec.Scope == "Namespaced" {
				namespaced = true
			}
			return _crd.Spec.Names.Plural, _crd.Spec.Group, _crd.Spec.Names.Singular, namespaced, nil
		}
		klog.V(5).Info("INFO ", fmt.Sprintf("Alias \"%s\" not matched in bundle CRDs.", alias))
	}
	klog.V(4).Info("INFO ", fmt.Sprintf("No customResource found with name or alias \"%s\" in path: \"%s\".", alias, crdsPath))
	// ~/.omc/customresourcedefinitions is an always-on fallback: it covers
	// resources whose CRDs are absent from the must-gather, which is exactly
	// when the fallback is needed most.
	home, _ := os.UserHomeDir()
	omcCrdsPath := home + "/.omc/customresourcedefinitions/"
	if ok, _ := Exists(omcCrdsPath); ok {
		crds, rErr := ReadDirForResources(omcCrdsPath)
		if rErr != nil {
			fmt.Fprintln(os.Stderr, rErr)
		}
		for _, f := range crds {
			if f.IsDir() {
				continue
			}
			crdYamlPath := omcCrdsPath + f.Name()
			crdByte, _ := ioutil.ReadFile(crdYamlPath)
			_crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal([]byte(crdByte), &_crd); err != nil {
				continue
			}
			if strings.Contains(alias, ".") {
				split := strings.Split(alias, ".")
				if len(split) > 1 {
					group := strings.Join(split[1:], ".")
					if !strings.HasPrefix(_crd.Spec.Group, group) {
						continue
					} else {
						_alias := strings.Join(split[:1], ".")
						if strings.ToLower(_crd.Spec.Names.Plural) == _alias || strings.ToLower(_crd.Spec.Names.Singular) == _alias || StringInSlice(_alias, _crd.Spec.Names.ShortNames) {
							namespaced := false
							if _crd.Spec.Scope == "Namespaced" {
								namespaced = true
							}
							aliasCache[strings.ToLower(_crd.Spec.Names.Kind)+"."+_crd.Spec.Group] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
							return _crd.Spec.Names.Plural, _crd.Spec.Group, _crd.Spec.Names.Singular, namespaced, nil
						}
					}
				}
			}
			aliasCache[strings.ToLower(_crd.Spec.Names.Kind)+"."+_crd.Spec.Group] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
			if strings.ToLower(_crd.Spec.Names.Kind) == alias || strings.ToLower(_crd.Spec.Names.Plural) == alias || strings.ToLower(_crd.Spec.Names.Singular) == alias || StringInSlice(alias, _crd.Spec.Names.ShortNames) || _crd.Spec.Names.Singular+"."+_crd.Spec.Group == alias {
				aliasCache[alias] = apiextensionsv1.CustomResourceDefinition{Spec: _crd.Spec}
				klog.V(4).Info("INFO ", fmt.Sprintf("Alias  \"%s\" found in path \"%s\".", alias, crdYamlPath))
				namespaced := false
				if _crd.Spec.Scope == "Namespaced" {
					namespaced = true
				}
				return _crd.Spec.Names.Plural, _crd.Spec.Group, _crd.Spec.Names.Singular, namespaced, nil
			}
			klog.V(5).Info("INFO ", fmt.Sprintf("Alias \"%s\" not found in path \"%s\".", alias, crdYamlPath))
		}
	}
	klog.V(4).Info("INFO ", fmt.Sprintf("No customResource found with name or alias \"%s\" in path: \"%s\".", alias, omcCrdsPath))
	return alias, "", "", false, fmt.Errorf("No customResource found with name or alias \"%s\".", alias)
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

func ReadDirForResources(path string) ([]os.DirEntry, error) {
	klog.V(5).Info("INFO ", fmt.Sprintf("opening '%s'\n", path))
	return readDirForResources(os.DirFS(path))
}

// readdir wraps around fs.ReadDir and only return valid resource yaml files
func readDirForResources(in fs.FS) ([]os.DirEntry, error) {
	resources := make([]os.DirEntry, 0)
	files, err := fs.ReadDir(in, ".")
	if err == nil {
		for _, file := range files {
			fileName := file.Name()
			// validate filename as per k8s validation of a resource
			if len(validation.IsDNS1123Subdomain(fileName)) == 0 {
				// only dirs or yaml files are expected as valid resources, e.g.:
				//  router-default-abcde12345-fgh678 or rendered-worker-abcdef123456.yaml
				if filepath.Ext(fileName) == ".yaml" || file.IsDir() {
					fInfo, _ := file.Info()
					// ignore empty files
					if fInfo.Size() > 0 {
						resources = append(resources, file)
					}
				}
			}
		}
	} else {
		err = fmt.Errorf("failed to read dir: %s", err)
	}
	return resources, err
}
