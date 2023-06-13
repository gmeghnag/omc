package tablegenerator

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/printers"

	//"github.com/gmeghnag/vars/types"

	helpers "github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func InternalResourceTable(runtimeObject runtime.Object, unstruct *unstructured.Unstructured) (*metav1.Table, error) {
	resourceKind := strings.ToLower(unstruct.GetKind())
	table, err := vars.TableGenerator.GenerateTable(runtimeObject, printers.GenerateOptions{Wide: vars.Wide, NoHeaders: false})
	if err != nil {
		return InternalUnstructuredApiResource(*unstruct)
	}
	for i, column := range table.ColumnDefinitions {
		if column.Name == "Age" {
			table.Rows[0].Cells[i] = helpers.TranslateTimestamp(unstruct.GetCreationTimestamp())
			if unstruct.GetKind() != "Node" {
				break
			}
		}
		if column.Name == "Roles" {
			var NodeRoles []string
			for i := range unstruct.GetLabels() {
				if strings.HasPrefix(i, "node-role.kubernetes.io/") {
					NodeRoles = append(NodeRoles, strings.Split(i, "/")[1])
				}
			}
			sort.Strings(NodeRoles)
			if len(NodeRoles) > 0 {
				table.Rows[0].Cells[i] = strings.Join(NodeRoles, ",")
			}

		}
	}
	if table.ColumnDefinitions[0].Name == "Name" {
		if vars.ShowKind == true {
			table.Rows[0].Cells[0] = resourceKind + "/" + unstruct.GetName()
		} else {
			table.Rows[0].Cells[0] = unstruct.GetName()
		}
	}

	if vars.ShowNamespace {
		table.ColumnDefinitions = append([]metav1.TableColumnDefinition{{Format: "string", Name: "Namespace"}}, table.ColumnDefinitions...)
		table.Rows[0].Cells = append([]interface{}{unstruct.GetNamespace()}, table.Rows[0].Cells...)
	}
	if vars.ShowLabelsBoolVar {
		table.ColumnDefinitions = append(table.ColumnDefinitions, metav1.TableColumnDefinition{Format: "string", Name: "Labels"})
		labels := helpers.ExtractLabels(unstruct.GetLabels())
		table.Rows[0].Cells = append(table.Rows[0].Cells, labels)
	}
	return table, err
}

func InternalUnstructuredApiResource(unstruct unstructured.Unstructured) (*metav1.Table, error) {
	resourceKind := strings.ToLower(unstruct.GetKind())
	table := &metav1.Table{}
	if vars.ShowNamespace == true && unstruct.GetNamespace() != "" {
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Namespace", Type: "string", Format: "name"},
			{Name: "Name", Type: "string", Format: "string"},
			{Name: "Created At", Type: "date"},
		}
		if vars.ShowKind == true || vars.Namespace == "" {
			table.Rows = []metav1.TableRow{{Cells: []interface{}{unstruct.GetNamespace(), resourceKind + "." + strings.Split(unstruct.GetAPIVersion(), "/")[0] + "/" + unstruct.GetName(), unstruct.GetCreationTimestamp().Time.UTC().Format("2006-01-02T15:04:05")}}}
		} else {
			table.Rows = []metav1.TableRow{{Cells: []interface{}{unstruct.GetNamespace(), unstruct.GetName(), unstruct.GetCreationTimestamp().Time.UTC().Format("2006-01-02T15:04:05")}}}
		}

	} else {
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Created At", Type: "date"},
		}
		if vars.ShowKind == true || vars.Namespace == "" {
			table.Rows = []metav1.TableRow{{Cells: []interface{}{resourceKind + "." + strings.Split(unstruct.GetAPIVersion(), "/")[0] + "/" + unstruct.GetName(), unstruct.GetCreationTimestamp().Time.UTC().Format("2006-01-02T15:04:05")}}}

		} else {
			table.Rows = []metav1.TableRow{{Cells: []interface{}{unstruct.GetName(), unstruct.GetCreationTimestamp().Time.UTC().Format("2006-01-02T15:04:05")}}}
		}

	}
	return table, nil
}

func GenerateCustomResourceTable(unstruct unstructured.Unstructured) (*metav1.Table, error) {
	resourceKind := strings.ToLower(unstruct.GetKind())
	table := &metav1.Table{}
	vars.CRD = nil
	crd, ok := vars.AliasToCrd[resourceKind+"."+strings.Split(unstruct.GetAPIVersion(), "/")[0]]
	if ok {
		vars.CRD = &apiextensionsv1.CustomResourceDefinition{Spec: crd.Spec}
	}

	cells := []interface{}{}
	if vars.ShowKind == true {
		if vars.ShowNamespace && unstruct.GetNamespace() != "" {
			table.ColumnDefinitions = []metav1.TableColumnDefinition{{Name: "Namespace", Format: "string"}, {Name: "Name", Format: "name"}}
			cells = []interface{}{unstruct.GetNamespace(), resourceKind + "/" + unstruct.GetName()}
		} else {
			table.ColumnDefinitions = []metav1.TableColumnDefinition{{Name: "Name", Format: "name"}}
			cells = []interface{}{resourceKind + "/" + unstruct.GetName()}
		}
	} else {
		if vars.ShowNamespace && unstruct.GetNamespace() != "" {
			table.ColumnDefinitions = []metav1.TableColumnDefinition{{Name: "Namespace", Format: "string"}, {Name: "Name", Format: "name"}}
			cells = []interface{}{unstruct.GetNamespace(), unstruct.GetName()}
		} else {
			table.ColumnDefinitions = []metav1.TableColumnDefinition{{Name: "Name", Format: "name"}}
			cells = []interface{}{unstruct.GetName()}
		}
	}
	for i, column := range vars.CRD.Spec.Versions {
		if (vars.CRD.Spec.Group + "/" + column.Name) == unstruct.GetAPIVersion() {
			if len(vars.CRD.Spec.Versions[i].AdditionalPrinterColumns) > 0 {
				for _, column := range vars.CRD.Spec.Versions[i].AdditionalPrinterColumns {
					table.ColumnDefinitions = append(table.ColumnDefinitions, metav1.TableColumnDefinition{Name: column.Name, Format: "string"})
					if column.Name == "Age" || column.Type == "date" {
						cells = append(cells, helpers.TranslateTimestamp(unstruct.GetCreationTimestamp()))
					} else {
						v := helpers.GetFromJsonPath(unstruct.Object, fmt.Sprintf("%s%s%s", "{", column.JSONPath, "}"))
						cells = append(cells, v)
					}
				}
			} else {
				table.ColumnDefinitions = append(table.ColumnDefinitions, metav1.TableColumnDefinition{Name: "Age", Format: "string"})
				cells = append(cells, helpers.TranslateTimestamp(unstruct.GetCreationTimestamp()))
			}
			break
		}
	}
	table.Rows = []metav1.TableRow{{Cells: cells}}
	if vars.ShowLabelsBoolVar {
		table.ColumnDefinitions = append(table.ColumnDefinitions, metav1.TableColumnDefinition{Format: "string", Name: "Labels"})
		labels := helpers.ExtractLabels(unstruct.GetLabels())
		table.Rows[0].Cells = append(table.Rows[0].Cells, labels)
	}

	return table, nil
}

func kind(resource *unstructured.Unstructured) string {
	longKind := strings.Split(resource.GetAPIVersion(), "/")[0]
	splittedKind := strings.Split(longKind, ".")
	return splittedKind[0]
}
