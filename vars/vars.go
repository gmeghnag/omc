package vars

import (
	"github.com/gmeghnag/omc/configpath"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"
)

var Namespace, MustGatherRootPath, OutputStringVar, Id, Container, OMCVersionHash, OMCVersionTag, DiffCmd, DefaultProject, ForResource string

// ConfigPathResolver is the global config path resolver
var ConfigPathResolver *configpath.Resolver
var AllNamespaceBoolVar bool

var EventTypes []string
var KnownResources map[string]map[string]interface{}
var TableGenerator *printers.HumanReadableGenerator

var Schema *runtime.Scheme
