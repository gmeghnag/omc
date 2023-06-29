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
package certs

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/cert"
)

const (
	missingKeyMsg           = "%s does not contain a key '%s'\n"
	missingCaContentMsg     = "%s missing content for key '%s'\n"
	parseFailureMsg         = "Failed to parse %s/%s : %v\n"
	objConvertionFailureMsg = "Failed to convert to %s: %s"
)

func getSupportedCaKeyNames() []string {
	return []string{"ca-bundle.crt", "ca.crt", "service-ca.crt"}
}

var Inspect = &cobra.Command{
	Use:   "inspect",
	Short: "certificate inspect",
	Run: func(cmd *cobra.Command, args []string) {
		resourceTypes := []string{"cm", "secret", "csr"}
		if len(args) == 1 {
			resourceTypes = strings.Split(strings.ToLower(args[0]), ",")
		}
		inspectResources(resourceTypes)
	},
}

type CertDetail struct {
	*unstructured.Unstructured
	CertType string
	*x509.Certificate
}

func NewCertDetail(obj *unstructured.Unstructured, certType string, certificate *x509.Certificate) *CertDetail {
	if certificate == nil {
		return &CertDetail{obj, certType, &x509.Certificate{}}
	}
	return &CertDetail{obj, certType, certificate}
}

// In case a nil certificate (cm/secret/csr not containing a certificate) was provided,
// forward x509.certificate.IsZero as an indication of a nil-value certificate
func (c CertDetail) IsZero() bool {
	return c.NotBefore.IsZero()
}

// use ValidFrom/To to prevent name collision with certificate.NotBefore/NotAfter
func (c CertDetail) ValidFrom() string {
	if c.NotBefore.IsZero() {
		return ""
	}
	return c.NotBefore.UTC().String()
}

// use ValidFrom/To to prevent name collision with certificate.NotBefore/NotAfter
func (c CertDetail) ValidTill() string {
	if c.NotAfter.IsZero() {
		return ""
	}
	return c.NotAfter.UTC().String()
}

func (c CertDetail) ValidFor() []string {
	validServingNames := []string{}
	if c.IsZero() {
		return validServingNames
	}
	for _, ip := range c.IPAddresses {
		validServingNames = append(validServingNames, ip.String())
	}
	for _, dnsName := range c.DNSNames {
		validServingNames = append(validServingNames, dnsName)
	}
	return validServingNames
}

func (c CertDetail) issuer() string {
	if c.IsZero() {
		return ""
	}
	if c.Subject.CommonName == c.Issuer.CommonName {
		return "[self]"
	}
	return c.Issuer.CommonName
}

func (c CertDetail) Usages() []string {
	usages := []string{}
	if c.IsZero() {
		return usages
	}
	for _, curr := range c.ExtKeyUsage {
		if curr == x509.ExtKeyUsageClientAuth {
			usages = append(usages, "client")
			continue
		}
		if curr == x509.ExtKeyUsageServerAuth {
			usages = append(usages, "serving")
			continue
		}

		usages = append(usages, fmt.Sprintf("%d", curr))
	}
	return usages
}

func (c CertDetail) subject() string {
	if c.IsZero() {
		return ""
	}
	return c.Subject.String()
}

func (c CertDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Namespace    string   `json:"namespace"`
		Name         string   `json:"name"`
		Kind         string   `json:"kind"`
		CertType     string   `json:"certType"`
		Subject      string   `json:"subject"`
		NotBefore    string   `json:"notBefore"`
		NotAfter     string   `json:"notAfter"`
		ValidFor     []string `json:"validFor,omitempty"`
		Issuer       string   `json:"issuer"`
		Groups       []string `json:"groups,omitempty"`
		Usages       []string `json:"usages,omitempty"`
		CreationDate string   `json:"creationDate"`
	}{
		Namespace:    c.GetNamespace(),
		Name:         c.GetName(),
		Kind:         c.GetKind(),
		CertType:     c.CertType,
		Subject:      c.Subject.String(),
		NotBefore:    c.ValidFrom(),
		NotAfter:     c.ValidTill(),
		ValidFor:     c.ValidFor(),
		Issuer:       c.issuer(),
		Groups:       c.Subject.Organization,
		Usages:       c.Usages(),
		CreationDate: c.GetCreationTimestamp().String(),
	})
}

func printParseFailure(w io.Writer, f string) {
	if showParseFailure {
		fmt.Fprintf(w, f)
	}
}

func inspectResources(resourceTypes []string) {
	var data [][]string
	var resources []*CertDetail
	_headers := []string{"namespace", "name", "kind", "age", "certtype", "subject", "notbefore", "notafter", "validfor", "issuer", "groups", "usages"}
	for _, resourceType := range resourceTypes {
		switch resourceType {
		case "cm", "configmap", "configmaps":
			var configmaps []*unstructured.Unstructured
			GetConfigMaps(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &configmaps)
			for _, r := range configmaps {
				resources = append(resources, inspectConfigMap(os.Stdout, r)...)
			}
		case "secret", "secrets":
			var secrets []*unstructured.Unstructured
			GetSecrets(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &secrets)
			for _, r := range secrets {
				resources = append(resources, inspectSecret(os.Stdout, r)...)
			}
		case "csr", "certificatesigningrequest", "certificatesigningrequests":
			var csrs []unstructured.Unstructured
			GetCertificateSigningRequests(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &csrs)
			for _, r := range csrs {
				resources = append(resources, inspectCSR(os.Stdout, &r)...)
			}
		}
	}
	for _, curr := range resources {
		age := helpers.GetAge(vars.MustGatherRootPath+"/cluster-scoped-resources/", curr.GetCreationTimestamp())
		_list := []string{
			curr.GetNamespace(),
			curr.GetName(),
			curr.GetKind(),
			age,
			curr.CertType,
			curr.subject(),
			curr.ValidTill(),
			curr.ValidFrom(),
			strings.Join(curr.ValidFor(), ","),
			curr.issuer(),
			strings.Join(curr.Subject.Organization, ","),
			strings.Join(curr.Usages(), ","),
		}
		data = helpers.GetData(data, vars.AllNamespaceBoolVar, false, "", vars.OutputStringVar, 8, _list)
	}
	helpers.PrintOutput(resources, 8, vars.OutputStringVar, "", vars.AllNamespaceBoolVar, false, _headers, data, "")
}

func inspectConfigMap(w io.Writer, obj *unstructured.Unstructured) []*CertDetail {
	resourceString := fmt.Sprintf("configmaps/%s[%s]", obj.GetName(), obj.GetNamespace())
	var cm corev1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &cm)
	if err != nil {
		fmt.Fprintf(w, objConvertionFailureMsg, obj.GetKind(), err)
	}
	var certdetails []*CertDetail
	for _, caKeyName := range getSupportedCaKeyNames() {
		caBundle, ok := cm.Data[caKeyName]
		if !ok {
			printParseFailure(w, fmt.Sprintf(missingKeyMsg, resourceString, caKeyName))
			continue
		}
		if len(caBundle) == 0 {
			printParseFailure(w, fmt.Sprintf(missingCaContentMsg, resourceString, caKeyName))
			continue
		}
		certificates, err := cert.ParseCertsPEM([]byte(caBundle))
		if err != nil {
			printParseFailure(w, fmt.Sprintf(parseFailureMsg, obj.GetKind(), obj.GetName(), err))
		}
		for _, cert := range certificates {
			certdetails = append(certdetails, NewCertDetail(obj, "ca-bundle", cert))
		}
	}

	// in case no valid data keys are found but we want to list resources anyway
	if listNonCerts && len(certdetails) == 0 {
		certdetails = append(certdetails, NewCertDetail(obj, "N/A", nil))
	}
	return certdetails
}

func inspectSecret(w io.Writer, obj *unstructured.Unstructured) []*CertDetail {
	resourceString := fmt.Sprintf("secret/%s[%s]", obj.GetName(), obj.GetNamespace())
	var secret corev1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &secret)
	if err != nil {
		fmt.Fprintf(w, objConvertionFailureMsg, obj.GetKind(), err)
	}
	tlsCrt, isTLS := secret.Data["tls.crt"]
	var certdetails []*CertDetail
	var isCA bool
	if isTLS {
		if len(tlsCrt) == 0 {
			printParseFailure(w, fmt.Sprintf(missingCaContentMsg, resourceString, "tls.crt"))
		}

		certificates, err := cert.ParseCertsPEM([]byte(tlsCrt))
		if err != nil {
			printParseFailure(w, fmt.Sprintf(parseFailureMsg, obj.GetKind(), obj.GetName(), err))
		}
		for _, cert := range certificates {
			certdetails = append(certdetails, NewCertDetail(obj, "certificate", cert))
		}
	} else {
		printParseFailure(w, fmt.Sprintf(missingKeyMsg, resourceString, "tls.crt"))
	}

	for _, caKeyName := range getSupportedCaKeyNames() {
		caBundle, ok := secret.Data[caKeyName]
		if !ok {
			printParseFailure(w, fmt.Sprintf(missingKeyMsg, resourceString, caKeyName))
			continue
		}
		if len(caBundle) == 0 {
			printParseFailure(w, fmt.Sprintf(missingCaContentMsg, resourceString, caKeyName))
			continue
		}
		isCA = true
		certificates, err := cert.ParseCertsPEM([]byte(caBundle))
		if err != nil {
			printParseFailure(w, fmt.Sprintf(parseFailureMsg, obj.GetKind(), obj.GetName(), err))
		}
		for _, cert := range certificates {
			certdetails = append(certdetails, NewCertDetail(obj, "ca-bundle", cert))
		}
	}
	if listNonCerts && len(certdetails) == 0 {
		certdetails = append(certdetails, NewCertDetail(obj, "N/A", nil))
	}

	if !isTLS && !isCA {
		printParseFailure(w, fmt.Sprintf("%s NOT a tls secret or token secret\n", resourceString))
	}
	return certdetails
}

func inspectCSR(w io.Writer, obj *unstructured.Unstructured) []*CertDetail {
	var certdetails []*CertDetail
	resourceString := fmt.Sprintf("secret/%s[%s]", obj.GetName(), obj.GetNamespace())
	var csr certificatesv1.CertificateSigningRequest
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &csr)
	if err != nil {
		fmt.Fprintf(w, objConvertionFailureMsg, obj.GetKind(), err)
	}
	if len(csr.Status.Certificate) == 0 {
		printParseFailure(w, fmt.Sprintf("%s NOT SIGNED\n", resourceString))
		if listNonCerts {
			certdetails = append(certdetails, NewCertDetail(obj, "csr", nil))
		}
	}

	certificates, err := cert.ParseCertsPEM([]byte(csr.Status.Certificate))
	if err != nil {
		printParseFailure(w, fmt.Sprintf(parseFailureMsg, obj.GetKind(), obj.GetName(), err))
		if listNonCerts {
			certdetails = append(certdetails, NewCertDetail(obj, "csr", nil))
		}
	}
	for _, cert := range certificates {
		certdetails = append(certdetails, NewCertDetail(obj, "ca-bundle", cert))
	}
	return certdetails
}
