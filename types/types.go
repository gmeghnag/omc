package types

import (
	"k8s.io/client-go/kubernetes"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	// "k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/apimachinery/pkg/api/meta"
	//runtime "k8s.io/apimachinery/pkg/runtime"
	//utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	_ "embed"
	//core "k8s.io/kubernetes/pkg/apis/core"
	//ocpinternal "github.com/openshift/openshift-apiserver/pkg/apps/printers/internalversion"
	// cliprint "k8s.io/cli-runtime/pkg/printers"
)

type Context struct {
	Id      string `json:"id"`
	Path    string `json:"path"`
	Current string `json:"current"`
	Project string `json:"project"`
}

type Config struct {
	Id             string    `json:"id,omitempty"`
	Contexts       []Context `json:"contexts,omitempty"`
	UseLocalCRDs   bool      `json:"use_local_crds,omitempty"`
	DiffCmd        string    `json:"diff_command,omitempty"`
	DefaultProject string    `json:"default_project,omitempty"`
}

type DescribeClient struct {
	Namespace string
	kubernetes.Interface
}

type UnstructuredList struct {
	ApiVersion string                      `json:"apiVersion"`
	Kind       string                      `json:"kind"`
	Items      []unstructured.Unstructured `json:"items"`
}

type JsonPathList struct {
	ApiVersion string                   `json:"apiVersion"`
	Kind       string                   `json:"kind"`
	Items      []map[string]interface{} `json:"items"`
}
