package get

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"
)

func TestHandleEmptyWideOutput(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		rtype     []string
		resources *types.UnstructuredList
		want      string
	}{
		{
			name:      "single cluster scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single cluster scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single namespaced scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakenamespacescopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "single namespaced scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakenamespacescopedresources.operator.openshift.io found in default namespace.\n",
		},
		{
			name:      "cluster and namespaced scoped crd all namespaces",
			namespace: "",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io found.\n",
		},
		{
			name:      "cluster and namespaced scoped crd default namespace",
			namespace: "default",
			rtype:     []string{"fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io"},
			want:      "No resources fakeclusterscopedresources.operator.openshift.io,fakenamespacescopedresources.operator.openshift.io found.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			vars.MustGatherRootPath = "../../testdata/"
			vars.Namespace = tt.namespace
			validateArgs(tt.rtype)
			handleOutput(&stdout, &stderr)
			if !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("Got: %v \n", stderr.String())
				t.Errorf("Want: %v \n", tt.want)
			}
			vars.GetArgs = make(map[string]map[string]struct{})
		})
	}
}
