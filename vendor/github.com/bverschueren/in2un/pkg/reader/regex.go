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
package reader

import (
	"regexp"

	"github.com/bverschueren/in2un/pkg/helpers"
	log "github.com/sirupsen/logrus"
)

type IRegex interface {
	Build() string
	// Do the regex on the incoming string
	// return true to stop or false to continue after a match
	// return matching string result
	Do(string) (bool, string)
}

func NewRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &Regex{
		resourceGroup: resourceGroup,
		resourceName:  resourceName,
		namespace:     namespace,
	}
}

type Regex struct {
	resourceGroup, resourceName, namespace string
}

func (r *Regex) Build() string {
	log.Tracef("Build in Regex")
	return ""
}

func (r *Regex) Do(in string) (bool, string) {
	return do(r, in)
}

type ConfigRegex struct {
	Regex
	base IRegex
}

// ConfigRegex is the "intermediate" regex: base->intermediate (e.g. ConditionalRegex) ->final
func NewConfigRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &ConfigRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: NewRegex(
			resourceGroup,
			resourceName,
			namespace,
		),
	}
}

func (c *ConfigRegex) Build() string {
	log.Tracef("Build in ConfigRegex")
	resourceGroupPart := helpers.Plural(c.resourceGroup)
	if c.resourceGroup == "all" {
		resourceGroupPart = `[a-z0-9\-]+`
	}
	// TODO: fix that `get storage` also matches storageclass resources
	reg := `^config(/storage)?/` + resourceGroupPart + `/`
	if helpers.Namespaced(c.resourceGroup) {
		if c.namespace != "" {
			if c.namespace == "_all_" {
				c.namespace = `[a-z0-9\-]+`
			}
			reg += c.namespace + `/`
		}
	}
	log.Tracef("got part '%s' in ConfigRegex", reg)
	return reg
}

type ConditionalRegex struct {
	Regex
	base IRegex
}

// ConditionalRegex is the "intermediate" regex: base->intermediate (e.g. ConditionalRegex) ->final
func NewConditionalRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &ConditionalRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: NewRegex(
			resourceGroup,
			resourceName,
			namespace,
		),
	}
}

func (c *ConditionalRegex) Build() string {
	log.Tracef("Build in ConditionalRegex")
	if c.namespace == "_all_" {
		c.namespace = `[a-z0-9\-]+`
	}
	reg := `^conditional/namespaces/` + c.namespace + `/` + helpers.Plural(c.resourceGroup) + `/`
	log.Tracef("got part '%s' in ConditionalRegex", reg)
	return reg
}

type OperatorConfigRegex struct {
	Regex
	base IRegex
}

func NewOperatorConfigRegex(resourceGroup, resourceName string) IRegex {
	return &OperatorConfigRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     "",
		},
		base: NewRegex(
			resourceGroup,
			resourceName,
			"",
		),
	}
}

// limit to /config/operatorconfig>.json
func (o *OperatorConfigRegex) Build() string {
	log.Tracef("Build in OperatorConfigRegex")
	resourceGroupPart := helpers.Plural(o.resourceGroup)
	reg := `^config(/storage)?/` + resourceGroupPart + `.json`
	log.Tracef("got part '%s' in OperatorConfigRegex", reg)
	return reg
}

func (c *OperatorConfigRegex) Do(in string) (bool, string) {
	log.Tracef("scanning '%s'", in)
	r := c.Build()
	log.Tracef("with '%s'", r)
	re := regexp.MustCompile(r)
	match := re.FindString(in)
	if match != "" {
		log.Tracef("found match '%s' for '%s' on %s\n", match, r, in)
		return true, match
	}
	return true, ""
}

type ResourceRegex struct {
	Regex
	base IRegex
}

// ResourceRegex is the "final" regex: base->intermediate (e.g. ConfigRegex) ->final
func NewResourceRegex(resourceGroup, resourceName, namespace string, base IRegex) IRegex {
	return &ResourceRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: base,
	}
}

func (r *ResourceRegex) Build() string {
	log.Tracef("Build in ResourceRegex")
	reg := r.base.Build()
	if r.resourceName != "" {
		// configmaps are expanded (namespace/cm-name/key) in insights, other resources are json (namespace/obj-name.json)
		reg += r.resourceName + `(.json|/[a-z0-9\-\.]+)$`
	} else {
		reg += `[a-z0-9\.\-]+(.json|/[a-z0-9\-\.]+)(.json)?$`
	}
	log.Tracef("got part '%s' in ResourceRegex", reg)
	return reg
}

func (c *ResourceRegex) Do(in string) (bool, string) {
	log.Tracef("Do in ResourceRegex")
	log.Tracef("scanning '%s'", in)
	r := c.Build()
	log.Tracef("with '%s'", r)
	re := regexp.MustCompile(r)
	match := re.FindString(in)
	if match != "" {
		if wellKnownInsightsJson(match) {
			log.Debugf("Found well-known path at '%s'\n", match)
			return true, match // stop processing next tokens
		}
		log.Tracef("found match '%s' for '%s' on %s\n", match, r, in)
		return false, match
	}
	return false, ""
}

type LogRegex struct {
	containerName string
	previous      bool
	Regex
}

func NewLogRegex(resourceGroup, resourceName, namespace, containerName string, previous bool) IRegex {
	return &LogRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		containerName: containerName,
		previous:      previous,
	}
}

func (r *LogRegex) Build() string {
	log.Tracef("Build in LogRegex")
	reg := ""
	// logs require a specific resourceName
	if r.resourceName == "" {
		return ""
	} else {
		if r.containerName == "" {
			reg += `logs/` + r.resourceName + `/[a-z0-9\.\-]+`
		} else {
			reg += `logs/` + r.resourceName + `/` + r.containerName
		}
	}
	if r.previous {
		reg += `_previous`
	} else {
		reg += `_current`
	}
	log.Tracef("got part '%s' in LogRegex", reg)
	return reg + `.log`
}

func (r *LogRegex) Do(in string) (bool, string) {
	return do(r, in)
}

type ResourceListRegex struct {
	Regex
}

func NewResourceListRegex() IRegex {
	return &ResourceListRegex{}
}

func (r *ResourceListRegex) Build() string {
	log.Tracef("Build in ResourceListRegex")
	return `^config(/storage)?/[a-z0-9]+(/|.json)`
}

func (r *ResourceListRegex) Do(in string) (bool, string) {
	return do(r, in)
}

func do(r IRegex, in string) (bool, string) {
	log.Tracef("do Regex with %+v", r)
	reg := r.Build()
	log.Tracef("scanning '%s'", in)
	log.Tracef("with '%s'", reg)
	re := regexp.MustCompile(reg)
	match := re.FindString(in)
	if match != "" {
		log.Tracef("found match '%s' for '%s' on %s\n", match, reg, in)
		return false, match
	}
	return false, ""
}
