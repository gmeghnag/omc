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
package deserializer

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var ErrUnknownResourcePath = fmt.Errorf("not a recognized path for a resource")

type ConfigMapData struct {
	// store ConfigMaps as a map of strings but wrap them in a unstructured.Unstructured
	// when calling Flatten() after reading all requested CM's
	// {"namespace": {
	// 	"configmap-name": {
	// 		"key1": "value1",
	// 		"key2": "value2",
	// 	}
	// }
	data collector
}

func NewConfigMapData() *ConfigMapData {
	return &ConfigMapData{
		data: make(collector),
	}
}

// {"namespace": {"name": {"key: "value"}}}
type collector = map[string]map[string]map[string]string

func (c *ConfigMapData) Upsert(namespace, name, key, value string) {
	object := make(map[string]string)
	if _, namespaceExists := c.data[namespace]; namespaceExists {
		if _, nameExists := c.data[namespace][name]; nameExists {
			object = c.data[namespace][name]
		}
		object[key] = value
		c.data[namespace][name] = object

	} else {
		c.data[namespace] = map[string]map[string]string{name: {key: value}}
	}
}

func (c *ConfigMapData) Flatten() []unstructured.Unstructured {
	out := []unstructured.Unstructured{}
	for namespace, nsContent := range c.data {
		for name := range nsContent {
			data := c.data[namespace][name]
			object := wrapConfigMap(name, namespace, data)
			out = append(out, *object)
		}
	}
	return out
}

// insights stores configmap data as plain files, so we re-construct them as configmaps and convert to unstructured again as per readResource api
func wrapConfigMap(name, namespace string, data map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.GetObjectKind().SetGroupVersionKind(u.GetObjectKind().GroupVersionKind())
	newObject := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
	}
	result, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&newObject)
	if err != nil {
		log.Fatal(err)
	}
	u.SetUnstructuredContent(result)
	return u
}

func configMapFromFilename(tarFilePath string) (name, namespace, key string, err error) {
	parts := strings.Split(strings.TrimSuffix(tarFilePath, "/"), "/")
	if len(parts) != 5 {
		return "", "", "", ErrUnknownResourcePath
	}
	return parts[3], parts[2], parts[4], nil
}
