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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

const haproxy_config_glob = "/ingress_controllers/*/*/haproxy.config"

var includeOpenShiftNamespaces bool

var Backends = &cobra.Command{
	Use:   "backends",
	Short: "Inspect haproxy configured backends.",
	Run: func(cmd *cobra.Command, args []string) {

		// in general if omc is not invoked with a specific --namespace / -n
		// option, it defaults to the user's current context project (see
		// root/root.go)
		// the approach for the `haproxy backends` subcommand
		// differs from omc's default behaviour as here we list backends for all
		// namespaces unless a specific namespace is provided through the root
		// flag
		var wantedNamespace string
		if cmd.Flags().Changed("namespace") {
			wantedNamespace = vars.Namespace
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintln(writer, "NAMESPACE\tNAME\tINGRESSCONTROLLER\tSERVICES\tPORT\tTERMINATION")
		for _, configfile := range haproxyConfigFiles(vars.MustGatherRootPath) {
			backends := parseHAProxyConfig(configfile, wantedNamespace)
			for _, b := range backends {
				fmt.Fprintln(writer, b)
			}
		}
		writer.Flush()
	},
}

func haproxyConfigFiles(root string) []string {
	pattern := root + haproxy_config_glob
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return files
}

// parse backend lines from a haproxy config file
// if a namespace is provided, only backends in that namespace are considered
func parseHAProxyConfig(filename string, wantedNamespace string) []*backend {
	ic := icFromFileName(filename)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var backends []*backend
	for scanner.Scan() {
		line := scanner.Text()
		backend := isBackendBlock(line)
		if backend != nil {
			if wantedNamespace == "" || backend.namespace == wantedNamespace {
				for scanner.Scan() {
					backendLine := scanner.Text()
					serverLine := isServerLine(backendLine)
					if serverLine != "" {
						backend.service = serviceFromServerLine(serverLine)
						backend.ingressController = ic
						break
					}
				}
				backends = append(backends, backend)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return backends
}

func icFromFileName(filename string) string {
	icRe := `ingress_controllers/([a-z0-9\-\_]*)/`
	re := regexp.MustCompile(icRe)

	matches := re.FindStringSubmatch(filename)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

type backend struct {
	termination, namespace, routeName, ingressController string
	service                                              *service
}

func (b backend) String() string {
	terminationType := func(s string) string {
		mapping := map[string]string{
			"be_edge_http": "edge/Redirect",
			"be_secure":    "reencrypt/Redirect",
			"be_tcp":       "passthrough/Redirect",
			"be_http":      "http",
		}
		return mapping[s]
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t", b.namespace, b.routeName, b.ingressController, b.service.serviceName, b.service.port, terminationType(b.termination))
}

func newBackendFromLine(raw []string) *backend {
	return &backend{
		termination: raw[1],
		namespace:   raw[2],
		routeName:   raw[3],
	}
}

func isBackendBlock(line string) *backend {
	backendRe := `^backend ([a-z0-9\-\_]*):([a-z0-9\-\_]*):([a-z0-9\-\_]*)$`
	re := regexp.MustCompile(backendRe)

	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil
	}

	if !includeOpenShiftNamespaces {
		matched, _ := regexp.MatchString(`openshift-.*`, matches[2])
		if matched {
			return nil
		}
	}
	return newBackendFromLine(matches)
}

type service struct {
	serviceName string
	port        *port
}

type port struct {
	portNr   int
	portName string
}

func (p port) String() string {
	if p.portName != "" {
		return fmt.Sprintf("%s(%d)", p.portName, p.portNr)
	}
	return fmt.Sprintf("%d", p.portNr)
}

// test if a line starts with '  server pod:' and return up up to the key/value as a string; empty string if no match
func isServerLine(line string) string {
	serverRe := `^  server pod:([a-z0-9\-\_\:\.]*) `
	re := regexp.MustCompile(serverRe)

	matches := re.FindStringSubmatch(line)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func serviceFromServerLine(line string) *service {
	parts := strings.Split(line, ":")
	portNr, err := strconv.Atoi(parts[4])
	if err != nil {
		fmt.Printf("Failed to convert port value (%+v) to an int.\n", parts[4])
		return &service{
			serviceName: parts[1],
			port:        &port{portName: parts[2]},
		}
	}

	return &service{
		serviceName: parts[1],
		port:        &port{portNr: portNr, portName: parts[2]},
	}
}
