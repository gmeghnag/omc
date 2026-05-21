package vars

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"
)

var Tail int64
var CfgFile, Namespace, MustGatherRootPath, OutputStringVar, LabelSelectorStringVar, Id, Container, OMCVersionHash, OMCVersionTag, DiffCmd, DefaultProject, ForResource string
var AllNamespaceBoolVar, ShowLabelsBoolVar, Previous, Rotated, AllContainers, UseLocalCRDs, SingleResource, Wide, ShowKind, ShowNamespace, ShowManagedFields, NoHeaders, InsecureLogs bool

var EventTypes []string
var GetArgs map[string]map[string]struct{}
var AliasToCrd map[string]apiextensionsv1.CustomResourceDefinition
var ArgPresent map[string]bool
var KnownResources map[string]map[string]interface{}
var TableGenerator *printers.HumanReadableGenerator
var CRD *apiextensionsv1.CustomResourceDefinition

var Schema *runtime.Scheme

var SortBy string
