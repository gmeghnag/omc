package vars

import (
	"bytes"

	"github.com/gmeghnag/omc/types"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"
)

var CfgFile, Namespace, MustGatherRootPath, OutputStringVar, LabelSelectorStringVar, Id, Container, OMCVersionHash, OMCVersionTag, DiffCmd, CurrentKind, LastKind, DefaultProject string
var AllNamespaceBoolVar, ShowLabelsBoolVar, Previous, AllContainers, UseLocalCRDs, SingleResource, Wide, ShowKind, ShowNamespace, ShowManagedFields, NoHeaders, InsecureLogs bool

var GetArgs map[string]map[string]struct{}
var AliasToCrd map[string]apiextensionsv1.CustomResourceDefinition
var ArgPresent map[string]bool
var KnownResources map[string]map[string]interface{}
var TableGenerator *printers.HumanReadableGenerator
var CRD *apiextensionsv1.CustomResourceDefinition

var Schema *runtime.Scheme

var UnstructuredList types.UnstructuredList
var JsonPathList types.JsonPathList

var Output bytes.Buffer

var Table metav1.Table
