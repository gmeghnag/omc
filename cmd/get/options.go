/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package get

// Options holds the configuration for a single get invocation. Fields are
// populated from the cobra-bound flags before Run is invoked.
type Options struct {
	RootPath string
	// RootPaths holds every must-gather root of the active context, ordered
	// most-recent-first. get reads from all of them and deduplicates by
	// metadata.uid (most recent wins). When empty it defaults to {RootPath},
	// so a single-must-gather context behaves exactly as before.
	RootPaths         []string
	Namespace         string
	NamespaceExplicit bool
	Output            string
	LabelSelector     string
	NoHeaders         bool
	AllNamespaces     bool
	ShowLabels        bool
	Wide              bool
	ShowKind          bool
	ShowNamespace     bool
	SortBy            string
	SingleResource    bool
	ShowManagedFields bool
	GetArgs           map[string]map[string]struct{}
	// GetArgsOrder lists the resolved "<plural>.<group>" keys in the order the
	// user requested them, so multi-resource output is deterministic (map
	// iteration order is not). Populated by validateArgs alongside GetArgs.
	GetArgsOrder []string
}

func newOptions() Options {
	return Options{
		GetArgs: make(map[string]map[string]struct{}),
	}
}

// roots returns the must-gather roots to read from, most-recent-first. It
// always returns at least one entry so callers can iterate unconditionally.
func (o *Options) roots() []string {
	if len(o.RootPaths) > 0 {
		return o.RootPaths
	}
	return []string{o.RootPath}
}
