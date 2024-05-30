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
	"reflect"
	"strings"
	"testing"
)

func TestNewCRILogFilter(t *testing.T) {
	tests := []struct {
		name      string
		levels    []string
		delimiter []byte
		expected  logLineFilter
	}{
		{
			name:      "New CRILogFilter is created with given parameters",
			levels:    []string{"info"},
			delimiter: []byte{' '},
			expected:  &CRILogFilter{[]string{"I"}, []byte{' '}},
		},
		{
			name:      "Empty filter levels return nil value CRILogFilter",
			levels:    nil,
			delimiter: nil,
			expected:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewCRILogFilter(tc.levels, tc.delimiter)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Expected : %v, got: %v", tc.expected, actual)
			}
		})
	}

}
func TestCRILogFilterLogLine(t *testing.T) {
	tests := []struct {
		name                    string
		input                   []byte
		levels                  []string
		expectedErrorText       string
		expectedOutputText      string
		expectedFilteredOutText string
	}{
		{
			// see filter.go: expeting a timestamp before the first delimiter
			name:                    "Error when timestamp not in expected field before first delimiter",
			input:                   []byte("2023-11-02T06:12:08.604741676Z"),
			levels:                  []string{"info", "warn", "error"},
			expectedErrorText:       "timestamp is not found",
			expectedOutputText:      "",
			expectedFilteredOutText: "",
		},
		{
			name:                    "Error when timestamp format is incorrect",
			input:                   []byte("2023-11-02T06: E1106 06:12:08.604741       1 test_app_go:542] My Error LogMessage"),
			levels:                  []string{"info", "warn", "error"},
			expectedErrorText:       "unexpected timestamp format",
			expectedOutputText:      "",
			expectedFilteredOutText: "",
		},
		{
			name:                    "Input contains filtered message",
			input:                   []byte("2023-11-02T06:12:08.604741676Z E1106 06:12:08.604741       1 test_app_go:542] My Error LogMessage"),
			levels:                  []string{"error"},
			expectedErrorText:       "",
			expectedOutputText:      "My Error LogMessage",
			expectedFilteredOutText: "",
		},
		{
			name:                    "Input does not contains filtered message",
			input:                   []byte("2023-11-02T06:12:08.604741676Z E1106 06:12:08.604741       1 test_app_go:542] My Error LogMessage"),
			levels:                  []string{"info"},
			expectedErrorText:       "",
			expectedOutputText:      "",
			expectedFilteredOutText: "My Error LogMessage",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCRILogFilter(tc.levels, nil)
			got, err := c.filterLogLine(tc.input)

			if tc.expectedErrorText != "" {
				if err == nil {
					t.Errorf("Expected error, got %v", string(got))
				} else if !strings.Contains(err.Error(), tc.expectedErrorText) {
					t.Errorf("Expected error '%v', got: '%v'", err.Error(), tc.expectedErrorText)
				}
			}
			if tc.expectedOutputText != "" {
				if !strings.Contains(string(got), tc.expectedOutputText) {
					t.Errorf("Expected output to contain '%v', got: '%v'", tc.expectedOutputText, string(got))
				}
			}
			if tc.expectedFilteredOutText != "" {
				if strings.Contains(string(got), tc.expectedFilteredOutText) {
					t.Errorf("Found '%v' while expecting to be filtered.\n", string(tc.expectedFilteredOutText))
				}
			}
		})
	}
}
