package vars

import (
	"github.com/gmeghnag/omc/configpath"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"
)

var Namespace, MustGatherRootPath, OutputStringVar, Id, Container, OMCVersionHash, OMCVersionTag, DiffCmd, DefaultProject, ForResource string

// MustGatherRootPaths holds every must-gather root of the active context,
// ordered most-recent-first. For a single-must-gather context it has one entry
// equal to MustGatherRootPath. Commands that merge across must-gathers (get) or
// resolve a resource to its most recent capture (describe, logs, events) read
// this; everything else keeps using MustGatherRootPath, which is always
// MustGatherRootPaths[0] (the most recent root).
var MustGatherRootPaths []string

// ConfigPathResolver is the global config path resolver
var ConfigPathResolver *configpath.Resolver
var AllNamespaceBoolVar bool

var EventTypes []string
var KnownResources map[string]map[string]interface{}
var TableGenerator *printers.HumanReadableGenerator

var Schema *runtime.Scheme
