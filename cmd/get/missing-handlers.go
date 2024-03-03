package get

import (
	"reflect"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	configv1 "github.com/openshift/api/config/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"k8s.io/kubernetes/pkg/printers"
)

func AddMissingHandlers(h printers.PrintHandler) {
	apiServiceColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name"},
		{Name: "Service", Type: "string"},
		{Name: "Available", Type: "string"},
		{Name: "Age", Type: "string"},
	}
	clusterVersionDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name"},
		{Name: "Version", Type: "string"},
		{Name: "Available", Type: "string"},
		{Name: "Progressing", Type: "string"},
		{Name: "Since", Type: "string"},
		{Name: "Status", Type: "string"},
	}

	customResourceDefinitionColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name"},
		{Name: "Ceated At", Type: "string"},
	}

	_ = h.TableHandler(apiServiceColumnDefinitions, printAPIService)
	_ = h.TableHandler(clusterVersionDefinitions, printClusterVersion)
	_ = h.TableHandler(customResourceDefinitionColumnDefinitions, printCustomResourceDefinitionv1)
	_ = h.TableHandler(customResourceDefinitionColumnDefinitions, printCustomResourceDefinitionv1beta1)
}

func printAPIService(obj *apiregistrationv1.APIService, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	service := "Local"
	if obj.Spec.Service != nil {
		service = obj.Spec.Service.Namespace + "/" + obj.Spec.Service.Name
	}
	available := "Unknown"
	for _, condition := range obj.Status.Conditions {
		if condition.Type == "Available" {
			available = string(condition.Status)
			if available != "True" {
				available = string(condition.Status) + " (" + condition.Reason + ")"
			}
			break
		}
	}
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.Name, service, available, "")
	return []metav1.TableRow{row}, nil
}

func printClusterVersion(obj *configv1.ClusterVersion, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	clusterOperatorName := obj.Name
	//version
	version := ""
	for _, h := range obj.Status.History {
		if h.State == "Completed" {
			version = h.Version
			break
		}
	}
	// conditions
	conditions := obj.Status.Conditions
	available := ""
	progressing := ""
	status := ""
	var lastsTransitionTime []v1.Time
	var lastTransitionTime v1.Time
	var zeroTime v1.Time
	for _, c := range conditions {
		//available
		if c.Type == "Available" {
			available = string(c.Status)
			lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
		}
		//progressing
		if c.Type == "Progressing" {
			progressing = string(c.Status)
			status = string(c.Message)
			lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
		}
		//status
		if c.Type == "Failing" {
			lastsTransitionTime = append(lastsTransitionTime, c.LastTransitionTime)
		}
	}
	//since
	for _, t := range lastsTransitionTime {
		if reflect.DeepEqual(lastTransitionTime, zeroTime) {
			lastTransitionTime = t
		} else {
			if t.Time.After(lastTransitionTime.Time) {
				lastTransitionTime = t
			}
		}
	}
	since := helpers.GetAge(vars.MustGatherRootPath, lastTransitionTime)

	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, clusterOperatorName, version, available, progressing, since, status)
	return []metav1.TableRow{row}, nil
}

func printCustomResourceDefinitionv1(obj *apiextensionsv1.CustomResourceDefinition, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	creationTimestamp := obj.GetCreationTimestamp()
	createdAt := creationTimestamp.UTC().Format(time.RFC3339Nano)
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.Name, createdAt)
	return []metav1.TableRow{row}, nil
}

func printCustomResourceDefinitionv1beta1(obj *apiextensionsv1beta1.CustomResourceDefinition, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	creationTimestamp := obj.GetCreationTimestamp()
	createdAt := creationTimestamp.UTC().Format(time.RFC3339Nano)
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.Name, createdAt)
	return []metav1.TableRow{row}, nil
}
