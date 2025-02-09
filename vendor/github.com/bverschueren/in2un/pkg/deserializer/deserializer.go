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
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const MissingTypeMetaFieldValue = "DUMMY"

type InsightsDeserializerOption func(*InsightsDeserializer)

type InsightsDeserializer struct {
	missingApiVersion, missingKind string
}

func NewInsightsDeserializer(o ...InsightsDeserializerOption) *InsightsDeserializer {
	result := &InsightsDeserializer{}
	for _, opt := range o {
		opt(result)
	}
	return result
}

func WithApiVersion(apiVersion string) InsightsDeserializerOption {
	return func(id *InsightsDeserializer) {
		if apiVersion == "" {
			apiVersion = MissingTypeMetaFieldValue
		}
		log.Debugf("overriding apiVersion with %s\n", apiVersion)
		id.missingApiVersion = apiVersion
	}
}

func WithKind(kind string) InsightsDeserializerOption {
	return func(id *InsightsDeserializer) {
		if kind == "" {
			kind = MissingTypeMetaFieldValue
		}
		log.Debugf("overriding kind with %s\n", kind)
		id.missingKind = kind
	}
}

func (id *InsightsDeserializer) JsonToUnstructed(raw []byte) (*unstructured.Unstructured, error) {
	result := &unstructured.Unstructured{}
	// First, try to unmarshal the raw json into an unstructured
	if err := result.UnmarshalJSON(raw); err != nil {
		// insights removes several typeMeta fields (Kind, apiVersion), eg:
		// https://github.com/openshift/insights-operator/blob/master/docs/insights-archive-sample/config/pod/openshift-insights/insights-operator-65bcbd8bbf-n5xcr.json
		// this causes unmarshal to fail, so try to insert fixup values for these fields and retry to unmarshal
		if fixed, fixErr := insertTypeMeta(raw, id.missingKind, id.missingApiVersion); fixErr != nil {
			return nil, fixErr
		} else {
			log.Trace("Trying to unmarshal after fixing missing TypeMeta fields")
			if retryErr := result.UnmarshalJSON(fixed); retryErr != nil {
				return nil, fmt.Errorf("error when trying to unmarshal fixed json into unstructured: %v", retryErr)
			}
			return result, nil
		}
	}
	return result, nil
}

func JsonToUnstructed(raw []byte) (*unstructured.Unstructured, error) {
	result := &unstructured.Unstructured{}
	// First, try to unmarshal the raw json into an unstructured
	if err := result.UnmarshalJSON(raw); err != nil {
		return nil, fmt.Errorf("error when trying to unmarshal into unstructured: %v", err)
	}
	return result, nil
}

// Certain resources are stripped off their Kind and apiVersion fields.
// In order to marshal those into Unstructured, insert dummy values for Kind and apiVersions or override with provided values
func insertTypeMeta(raw []byte, kind, apiVersion string) ([]byte, error) {
	var marshalled map[string]interface{}
	if err := json.Unmarshal(raw, &marshalled); err != nil {
		return nil, fmt.Errorf("error when trying to unmarshal json: %v", err)
	}
	if _, ok := marshalled["kind"]; !ok {
		if kind == "" {
			kind = MissingTypeMetaFieldValue
		}
		marshalled["kind"] = kind
	}
	if _, ok := marshalled["apiVersion"]; !ok {
		if apiVersion == "" {
			apiVersion = MissingTypeMetaFieldValue
		}
		marshalled["apiVersion"] = apiVersion
	}
	newData, jsonMarshalErr := json.Marshal(&marshalled)
	if jsonMarshalErr != nil {
		return nil, fmt.Errorf("error when trying to marshal json: %v", jsonMarshalErr)
	}
	return newData, nil
}
