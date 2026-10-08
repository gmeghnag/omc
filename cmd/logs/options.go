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
package logs

// Options holds the configuration for a single logs invocation. Fields are
// populated from the cobra-bound flags before Run is invoked.
type Options struct {
	RootPath string
	// RootPaths holds every must-gather root of the active context, ordered
	// most-recent-first. logs reads a pod from the most recent must-gather that
	// contains it (the same capture get shows); it never falls back to an older
	// capture's logs. Empty means a single-must-gather context (= {RootPath}).
	RootPaths     []string
	Namespace     string
	Container     string
	Previous      bool
	Rotated       bool
	AllContainers bool
	Insecure      bool
	Tail          int64
}
