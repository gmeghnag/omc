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
	"fmt"
	"github.com/gmeghnag/omc/cmd/get/certificate"
	"github.com/gmeghnag/omc/cmd/get/core"
	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
	"io"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"
	"os"
	"strings"
)

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

func inspectResources(resourceTypes []string) {
	for _, resourceType := range resourceTypes {
		switch resourceType {
		case "cm", "configmap", "configmaps":
			var configmaps []*corev1.ConfigMap
			core.GetConfigMaps(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &configmaps)
			for _, r := range configmaps {
				if err := certInspect(r); err != nil {
					fmt.Println(err)
				}
			}
		case "secret", "secrets":
			var secrets []*corev1.Secret
			core.GetSecrets(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &secrets)
			for _, r := range secrets {
				if err := certInspect(r); err != nil {
					fmt.Println(err)
				}
			}
		case "csr", "certificatesigningrequest", "certificatesigningrequests":
			var csrs []certificatesv1.CertificateSigningRequest
			certificate.GetCertificateSigningRequests(vars.MustGatherRootPath, vars.Namespace, "", vars.AllNamespaceBoolVar, &csrs)
			for _, r := range csrs {
				if err := certInspect(r); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func certInspect(resource interface{}) error {
	switch castObj := resource.(type) {
	case *corev1.ConfigMap:
		inspectConfigMap(os.Stdout, castObj)
	case *corev1.Secret:
		inspectSecret(os.Stdout, castObj)
	case *certificatesv1.CertificateSigningRequest:
		inspectCSR(os.Stdout, castObj)
	default:
		return fmt.Errorf("unhandled resource: %T", castObj)
	}
	return nil
}

func inspectConfigMap(w io.Writer, obj *corev1.ConfigMap) {
	resourceString := fmt.Sprintf("configmaps/%s[%s]", obj.Name, obj.Namespace)
	caBundle, ok := obj.Data["ca-bundle.crt"]
	if !ok {
		fmt.Fprintf(w, "%s NOT a ca-bundle\n", resourceString)
		return
	}
	if len(caBundle) == 0 {
		fmt.Fprintf(w, "%s MISSING ca-bundle content\n", resourceString)
		return
	}

	fmt.Fprintf(w, "%s - ca-bundle (%v)\n", resourceString, obj.CreationTimestamp.UTC())
	certificates, err := cert.ParseCertsPEM([]byte(caBundle))
	if err != nil {
		fmt.Fprintf(w, "    ERROR - %v\n", err)
		return
	}
	for _, curr := range certificates {
		fmt.Fprintf(w, "    %s\n", certDetail(curr))
	}
}

func inspectSecret(w io.Writer, obj *corev1.Secret) {
	resourceString := fmt.Sprintf("secrets/%s[%s]", obj.Name, obj.Namespace)
	tlsCrt, isTLS := obj.Data["tls.crt"]
	if isTLS {
		fmt.Fprintf(w, "%s - tls (%v)\n", resourceString, obj.CreationTimestamp.UTC())
		if len(tlsCrt) == 0 {
			fmt.Fprintf(w, "%s MISSING tls.crt content\n", resourceString)
			return
		}

		certificates, err := cert.ParseCertsPEM([]byte(tlsCrt))
		if err != nil {
			fmt.Fprintf(w, "    ERROR - %v\n", err)
			return
		}
		for _, curr := range certificates {
			fmt.Fprintf(w, "    %s\n", certDetail(curr))
		}
	}

	caBundle, isCA := obj.Data["ca.crt"]
	if isCA {
		fmt.Fprintf(w, "%s - token secret (%v)\n", resourceString, obj.CreationTimestamp.UTC())
		if len(caBundle) == 0 {
			fmt.Fprintf(w, "%s MISSING ca.crt content\n", resourceString)
			return
		}

		certificates, err := cert.ParseCertsPEM([]byte(caBundle))
		if err != nil {
			fmt.Fprintf(w, "    ERROR - %v\n", err)
			return
		}
		for _, curr := range certificates {
			fmt.Fprintf(w, "    %s\n", certDetail(curr))
		}
	}

	if !isTLS && !isCA {
		fmt.Fprintf(w, "%s NOT a tls secret or token secret\n", resourceString)
		return
	}
}

func inspectCSR(w io.Writer, obj *certificatesv1.CertificateSigningRequest) {
	resourceString := fmt.Sprintf("csr/%s", obj.Name)
	if len(obj.Status.Certificate) == 0 {
		fmt.Fprintf(w, "%s NOT SIGNED\n", resourceString)
		return
	}

	fmt.Fprintf(w, "%s - (%v)\n", resourceString, obj.CreationTimestamp.UTC())
	certificates, err := cert.ParseCertsPEM([]byte(obj.Status.Certificate))
	if err != nil {
		fmt.Fprintf(w, "    ERROR - %v\n", err)
		return
	}
	for _, curr := range certificates {
		fmt.Fprintf(w, "    %s\n", certDetail(curr))
	}
}

func certDetail(certificate *x509.Certificate) string {
	humanName := certificate.Subject.CommonName
	signerHumanName := certificate.Issuer.CommonName
	if certificate.Subject.CommonName == certificate.Issuer.CommonName {
		signerHumanName = "<self>"
	}

	usages := []string{}
	for _, curr := range certificate.ExtKeyUsage {
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

	validServingNames := []string{}
	for _, ip := range certificate.IPAddresses {
		validServingNames = append(validServingNames, ip.String())
	}
	for _, dnsName := range certificate.DNSNames {
		validServingNames = append(validServingNames, dnsName)
	}
	servingString := ""
	if len(validServingNames) > 0 {
		servingString = fmt.Sprintf(" validServingFor=[%s]", strings.Join(validServingNames, ","))
	}

	groupString := ""
	if len(certificate.Subject.Organization) > 0 {
		groupString = fmt.Sprintf(" groups=[%s]", strings.Join(certificate.Subject.Organization, ","))
	}

	return fmt.Sprintf("%q [%s]%s%s issuer=%q (%v to %v)", humanName, strings.Join(usages, ","), groupString, servingString, signerHumanName, certificate.NotBefore.UTC(), certificate.NotAfter.UTC())
}
