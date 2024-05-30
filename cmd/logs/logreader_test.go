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
	"compress/gzip"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

const testdata = "../../testdata/"

func TestRotatedFiles(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		expected *[]string
	}{
		{
			name:     "Find plain and gzipped log files in expected rotated directory",
			fixture:  testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs",
			expected: &[]string{"rotated/1.log.20231102-061208", "rotated/1.log.20231102-061208.gz"},
		},
		{
			name:     "Find no files if rotated directory does not exist",
			fixture:  testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/",
			expected: &[]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := rotatedFiles(tc.fixture)

			if len(*actual) != len(*tc.expected) {
				t.Fatalf("Expected : %v, got: %v", tc.expected, actual)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		expectedError bool
		expected      io.Reader
	}{
		{
			name:          "Error if file cannot be opened",
			fixture:       testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/missing.log",
			expectedError: true,
			expected:      nil,
		},
		{
			name:          "Return *os.File for plain file",
			fixture:       testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/current.log",
			expectedError: false,
			expected:      &os.File{},
		},
		{
			name:          "Return *gzip.Reader for gzipped file",
			fixture:       testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/rotated/1.log.20231102-061208.gz",
			expectedError: false,
			expected:      &gzip.Reader{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := open(tc.fixture)
			if err == nil && tc.expectedError {
				t.Errorf("Expected error, got %v", got)
			}
			expectedType := reflect.TypeOf(tc.expected)
			gotType := reflect.TypeOf(got)
			if expectedType != gotType {
				t.Errorf("Expected %v, got %v", expectedType, gotType)
			}
		})
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name                    string
		logReader               *LogReader
		expectedText            string
		expectedFilteredOutText string
	}{
		{
			name:                    "Read file unfiltered",
			logReader:               &LogReader{testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/", &[]string{"current.log"}, nil},
			expectedText:            "My Info LogMessage",
			expectedFilteredOutText: "",
		},
		{
			name:                    "Read file and apply filter to every line.",
			logReader:               &LogReader{testdata + "namespaces/test-namespace/pods/test-pod/test-container/test-container/logs/", &[]string{"current.log"}, &SimpleToUpperLogFilter{}},
			expectedText:            "LOGMESSAGE",
			expectedFilteredOutText: "LogMessage",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			tc.logReader.Read(output)
			if tc.expectedFilteredOutText != "" {
				if strings.Contains(output.String(), tc.expectedFilteredOutText) {
					t.Errorf("Found '%v' while expecting to be filtered.\n", tc.expectedFilteredOutText)
				}
			}
			if !strings.Contains(output.String(), tc.expectedText) {
				t.Errorf("Got: %v \n", output.String())
				t.Errorf("Want: %v \n", tc.expectedText)
			}
		})
	}
}

// dummy implementation of the logLineFilter interface for testing purposes
type SimpleToUpperLogFilter struct{}

func (s SimpleToUpperLogFilter) filterLogLine(log []byte) ([]byte, error) {
	return bytes.ToUpper(log), nil
}
