/*
Copyright © 2024 Bram Verschueren <bverschueren@redhat.com>

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
package helpers

import (
	"slices"

	log "github.com/sirupsen/logrus"
)

// ugly offline hack to check for namespacedness
// since we have no definitions, need to rely on a static list
func Namespaced(name string) bool {
	log.Debug("Using static list to check for namespacedness as best effort")
	var clusterScoped = []string{"machineconfig", "machineconfigpool", "clusteroperator", "node", "storageclass", "hostsubnet"}
	return !slices.Contains(clusterScoped, name)
}

func Plural(singular string) string {
	log.Debug("Using static plural parser as best effort")
	// add "es" for words ending in "ss" (e.g. storageclass)
	if string(singular[len(singular)-2:]) == `ss` {
		return singular + `es?`
	}
	// add ? for words already ending with s as this may indicate plural form already (e.g. pods)
	if string(singular[len(singular)-1:]) == `s` {
		return singular + `?`
	}
	return singular + `s?`
}

func Unalias(alias string) string {
	log.Debug("Using static alias map as best effort")
	aliases := map[string]string{
		"mc":  "machineconfig",
		"cm":  "configmap",
		"co":  "clusteroperator",
		"ns":  "namespace",
		"pv":  "persistentvolume",
		"pvc": "persistentvolumeclaim",
	}
	if unalias, ok := aliases[alias]; ok {
		return unalias
	}
	return alias
}
