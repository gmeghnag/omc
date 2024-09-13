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
package haproxy

import (
	"reflect"
	"testing"
)

const testdata = "../../testdata/"

func TestParseHAProxyConfig(t *testing.T) {
	tests := []struct {
		name, configFile, wantedNamespace string
		includeOpenShiftNamespaces        bool
		expected                          []*backend
	}{
		{
			name:                       "Parse HAProxy config and extract backends",
			configFile:                 "../../testdata/ingress_controllers/default/router-default-abc123-a1b1c3/haproxy.config",
			wantedNamespace:            "",
			includeOpenShiftNamespaces: false,
			expected: []*backend{
				&backend{namespace: "testdata", routeName: "rails-postgresql-example", ingressController: "default", service: &service{serviceName: "rails-postgresql-example", port: &port{portNr: 8080, portName: "web"}}, termination: "be_http"},
				&backend{namespace: "testdata", routeName: "app.example.com", ingressController: "default", service: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: ""}}, termination: "be_http"},
				&backend{namespace: "other-testdata", routeName: "hello-node-secure", ingressController: "default", service: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: ""}}, termination: "be_edge_http"}},
		},
		{
			name:                       "Parse HAProxy config and extract backends including openshift-*",
			configFile:                 "../../testdata/ingress_controllers/default/router-default-abc123-a1b1c3/haproxy.config",
			wantedNamespace:            "",
			includeOpenShiftNamespaces: true,
			expected: []*backend{
				&backend{namespace: "openshift-monitoring", routeName: "thanos-querier", ingressController: "default", service: &service{serviceName: "thanos-querier", port: &port{portNr: 9091, portName: "web"}}, termination: "be_secure"},
				&backend{namespace: "testdata", routeName: "rails-postgresql-example", ingressController: "default", service: &service{serviceName: "rails-postgresql-example", port: &port{portNr: 8080, portName: "web"}}, termination: "be_http"},
				&backend{namespace: "testdata", routeName: "app.example.com", ingressController: "default", service: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: ""}}, termination: "be_http"},
				&backend{namespace: "other-testdata", routeName: "hello-node-secure", ingressController: "default", service: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: ""}}, termination: "be_edge_http"}},
		},
		{
			name:                       "Parse HAProxy config and extract backends matching a namespace",
			configFile:                 "../../testdata/ingress_controllers/default/router-default-abc123-a1b1c3/haproxy.config",
			wantedNamespace:            "testdata",
			includeOpenShiftNamespaces: true,
			expected: []*backend{
				&backend{namespace: "testdata", routeName: "rails-postgresql-example", ingressController: "default", service: &service{serviceName: "rails-postgresql-example", port: &port{portNr: 8080, portName: "web"}}, termination: "be_http"},
				&backend{namespace: "testdata", routeName: "app.example.com", ingressController: "default", service: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: ""}}, termination: "be_http"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			includeOpenShiftNamespaces = tc.includeOpenShiftNamespaces
			found := parseHAProxyConfig(tc.configFile, tc.wantedNamespace)

			if !reflect.DeepEqual(found, tc.expected) {
				t.Fatalf("Expected : %+v, got: %+v", tc.expected, found)
			}
		})
	}
}

func TestHaproxyConfigFiles(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		expected []string
	}{
		{
			name:     "Find haproxy config files using glob pattern",
			root:     testdata,
			expected: []string{"./testdata/ingress_controllers/default/router-default-abc123-a1b1c3/haproxy.config", "./testdata/ingress_controllers/shard/router-default-xyz789-x7y8z9/haproxy.config"},
		},
		{
			name:     "Return empty slice if no config files found.",
			root:     testdata + "/fake",
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found := haproxyConfigFiles(tc.root)

			if len(found) != len(tc.expected) {
				t.Fatalf("Expected : %v, got: %v", tc.expected, found)
			}
		})
	}
}

func TestIcFromFileName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Find IngressController name from haproxy.config file location",
			filename: "./testdata/ingress_controllers/default/router-default-abc123-a1b1c3/haproxy.config",
			expected: "default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found := icFromFileName(tc.filename)

			if len(found) != len(tc.expected) {
				t.Fatalf("Expected : %v, got: %v", tc.expected, found)
			}
		})
	}
}

func TestIsBackendBlock(t *testing.T) {
	tests := []struct {
		name             string
		line             string
		includeOpenShift bool
		expected         *backend
	}{
		{
			name:             "return backend from valid backend block",
			line:             "backend be_edge_http:testdata:hello-node",
			includeOpenShift: true,
			expected:         &backend{termination: "be_edge_http", namespace: "testdata", routeName: "hello-node", ingressController: "", service: (*service)(nil)},
		},
		{
			name:             "return backend from valid backend block including dot in route",
			line:             "backend be_edge_http:testdata:hello-node.example.com",
			includeOpenShift: true,
			expected:         &backend{termination: "be_edge_http", namespace: "testdata", routeName: "hello-node.example.com", ingressController: "", service: (*service)(nil)},
		},
		{
			name:             "return nil from invalid backend block",
			line:             "nonbackend be_edge_http:testdata:hello-node",
			includeOpenShift: true,
			expected:         nil,
		},
		{
			name:             "return nil from valid backend block for openshift-managed route",
			line:             "backend be_edge_http:openshift-namespace:hello-node",
			includeOpenShift: false,
			expected:         nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			includeOpenShiftNamespaces = tc.includeOpenShift
			found := isBackendBlock(tc.line)

			if !reflect.DeepEqual(tc.expected, found) {
				t.Fatalf("Expected : %#v, got: %#v", tc.expected, found)
			}
		})
	}
}

func TestIsServerLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "return service from valid server line",
			line:     "  server pod:hello-node-595bfd9b77-4rm94:hello-node::10.129.2.15:8080 10.129.2.15:8080 cookie 863159b6f80f224951e08d6c052520a4 weight 1",
			expected: "hello-node-595bfd9b77-4rm94:hello-node::10.129.2.15:8080",
		},
		{
			name:     "return nil from invalid server line",
			line:     "nonserver po",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found := isServerLine(tc.line)

			if tc.expected != found {
				t.Fatalf("Expected : %#v, got: %#v", tc.expected, found)
			}
		})
	}
}

func TestServiceFromServerLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *service
	}{
		{
			name:     "return service from server line with named port",
			line:     "hello-node-595bfd9b77-4rm94:hello-node:web:10.129.2.15:8080",
			expected: &service{serviceName: "hello-node", port: &port{portNr: 8080, portName: "web"}},
		},
		{
			name:     "return service from server line without named port",
			line:     "hello-node-595bfd9b77-4rm94:hello-node::10.129.2.15:8080",
			expected: &service{serviceName: "hello-node", port: &port{portNr: 8080}},
		},
		{
			name:     "return service with portName only when portNumber is not an int",
			line:     "hello-node-595bfd9b77-4rm94:hello-node:web:10.129.2.15:eighthy-eighty",
			expected: &service{serviceName: "hello-node", port: &port{portName: "web"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found := serviceFromServerLine(tc.line)

			if !reflect.DeepEqual(tc.expected, found) {
				t.Fatalf("Expected : %#v, got: %#v", tc.expected, found)
			}
		})
	}
}

func TestServiceString(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "service.String() prints portName if included in server line",
			line:     "hello-node-595bfd9b77-4rm94:hello-node:web:10.129.2.15:8080",
			expected: "web(8080)",
		},
		{
			name:     "service.String() emits portName if missing from server line",
			line:     "hello-node-595bfd9b77-4rm94:hello-node::10.129.2.15:8080",
			expected: "8080",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found := serviceFromServerLine(tc.line)
			printed := found.port.String()

			if tc.expected != printed {
				t.Fatalf("Expected : %s, got: %s", tc.expected, printed)
			}
		})
	}
}
