package uget

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
)

type Column struct {
	Name        string `json:"name"`
	JSONPath    string `json:"jsonPath"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

type AdditionalColumns struct {
	Columns []Column `json:"columns,omitempty"`
}

func getFromJsonPath(data interface{}, jsonPathTemplate string) string {
	buf := new(bytes.Buffer)
	jPath := jsonpath.New("out")
	jPath.AllowMissingKeys(false)
	jPath.EnableJSONOutput(false)
	err := jPath.Parse(jsonPathTemplate)
	if err != nil {
		fmt.Println("error: error parsing jsonpath " + jsonPathTemplate + ", " + err.Error())
		os.Exit(1)
	}
	jPath.Execute(buf, data)
	return buf.String()
}

func toJsonPath(path string) string {
	path = strings.TrimPrefix(path, "{")
	path = strings.TrimSuffix(path, "}")
	path = "{" + path + "}"
	return path
}

type UnstrctList struct {
	ApiVersion string                      `json:"apiVersion"`
	Kind       string                      `json:"kind"`
	Items      []unstructured.Unstructured `json:"items"`
}

func matchKind(kinds []string, objectKind string) bool {
	if len(kinds) == 0 {
		return true
	}
	objectKind = strings.ToLower(objectKind)
	for _, kind := range kinds {
		if kind == objectKind {
			return true
		}
	}
	return false
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
