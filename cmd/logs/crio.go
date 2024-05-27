/*
Copyright © 2023 Bram Verschueren <bverschueren@redhat.com>

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

import (
	"bytes"
	"fmt"
	"time"

	"k8s.io/utils/strings/slices"
)

const (
	// RFC3339NanoFixed is the fixed width version of time.RFC3339Nano.
	RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"
	// RFC3339NanoLenient is the variable width RFC3339 time format for lenient parsing of strings into timestamps.
	RFC3339NanoLenient = "2006-01-02T15:04:05.999999999Z07:00"
	// timeFormatOut is the format for writing timestamps to output.
	timeFormatOut = RFC3339NanoFixed
	// timeFormatIn is the format for parsing timestamps from other logs.
	timeFormatIn = RFC3339NanoLenient
)

type CRILogFilter struct {
	levels    []string
	delimiter []byte
}

// create a new CRILogFilter or nil if no filter values are given
func NewCRILogFilter(wantedLevels []string, delimiter []byte) logLineFilter {
	var lf logLineFilter
	if delimiter == nil {
		delimiter = []byte{' '}
	}
	levels := []string{}
	for _, l := range wantedLevels {
		switch l {
		case "info":
			levels = append(levels, "I")
		case "warn":
			levels = append(levels, "W")
		case "error":
			levels = append(levels, "E")
		}
	}
	if len(levels) > 0 {
		lf = &CRILogFilter{levels, delimiter}
	}
	return lf
}

func (c *CRILogFilter) filterLogLine(log []byte) ([]byte, error) {
	var err error
	// Parse timestamp
	idx := bytes.Index(log, c.delimiter)
	if idx < 0 {
		return []byte{}, fmt.Errorf("timestamp is not found")
	}
	//only to check if timestamp is valid
	_, err = time.Parse(timeFormatIn, string(log[:idx]))
	if err != nil {
		return []byte{}, fmt.Errorf("unexpected timestamp format %q: %v", timeFormatIn, err)
	}

	// Parse stream type
	_log := log[idx+1:]
	idx = bytes.Index(_log, c.delimiter)
	if idx < 0 {
		idx = len(string(_log))
	}
	stream := string(_log[:idx])
	if len(stream) == 0 {
		return []byte{}, nil
	}
	if slices.Contains(c.levels, string(stream[0])) && isNumber(stream[1]) {
		return log, nil
	}

	return []byte{}, nil
}
