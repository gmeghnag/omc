package vars

import (
	"github.com/gmeghnag/omc/configpath"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"
)

var Namespace, MustGatherRootPath, OutputStringVar, Id, Container, OMCVersionHash, OMCVersionTag, DiffCmd, DefaultProject, ForResource string

// InstallMethod records the package manager that owns the running binary. It
// is empty for the release binaries published by this repo, which can upgrade
// themselves; packagers set it at build time, e.g. the homebrew formula builds
// with -ldflags "-X github.com/gmeghnag/omc/vars.InstallMethod=homebrew" so
// that "omc upgrade" defers to "brew upgrade" instead of replacing the binary.
var InstallMethod string

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
