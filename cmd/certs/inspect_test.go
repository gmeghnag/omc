package certs

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"testing"
)

const (
	pemServerCert = `-----BEGIN CERTIFICATE-----
MIICtDCCAZygAwIBAgIUNknYnqtlQSl7kpt7BrVp4EAjQyAwDQYJKoZIhvcNAQEL
BQAwJTEjMCEGA1UEAwwaa3ViZS1jc3Itc2lnbmVyX0AxMjM0NTY3ODkwHhcNMjMw
NDIxMTQzNDMzWhcNMjQwNDExMTQzNDMzWjBBMRUwEwYDVQQKDAxzeXN0ZW06bm9k
ZXMxKDAmBgNVBAMMH3N5c3RlbTpub2RlczpteW5vZGUuZXhhbXBsZS5jb20wWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAAQOWYlirUnQKAViuLVbcabWt2ikmYNB0XVA
TAznTzMN8h9zRNJyclXRiMO/esAitZC3j0BO2LstdwRjzT5Q3XNIo4GKMIGHMB8G
A1UdIwQYMBaAFAGbrjSLLirGWLWcJ3TsfqPf+6jmMAkGA1UdEwQCMAAwCwYDVR0P
BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMDcGA1UdEQQwMC6CH3N5c3RlbTpu
b2RlczpteW5vZGUuZXhhbXBsZS5jb22CCzE5Mi4xNjguMS4xMA0GCSqGSIb3DQEB
CwUAA4IBAQAXXpRgZ3gwiFMnYDsmG0h15915vTB8uX8jf/bdhCJWn7LlzJpLuGjl
o60jf/17tMLgp0l4vcfZDvK7RKWz48j2dfTxd0t82QeMkkD92Hm8yyaQur7XcwtP
cFWiMmVfLejhnQqBFQgrGMniWJhIN9uQUQul2ANKZ5pwmN8cRt6Oeb9Eg6yHzHUK
wlo4f3l3lPaihoLE7F3dk4eDBQTCypFQhoxJf+4OMft5fgcC25134MG2ShYiUA9t
ZQzAy9dZrLzbfQWZ5km/juWC7z3FgS+WNDTd76WrlzVuGW4qBg+0TJ/j4oLpzjrL
aXN6uc2j4F2o1XAZQdTkKYvTM4nnwe/+
-----END CERTIFICATE-----`
	pemCaCert = `-----BEGIN CERTIFICATE-----
MIICODCCAaGgAwIBAgIUeoMh/N7rDvcV4f3d9i+Y1/d95K4wDQYJKoZIhvcNAQEL
BQAwLjENMAsGA1UECgwEdGVzdDEMMAoGA1UECwwDT3JnMQ8wDQYDVQQDDAZSb290
Q0EwHhcNMjMwNDIxMDgzNjQxWhcNMjQwNDIwMDgzNjQxWjAuMQ0wCwYDVQQKDAR0
ZXN0MQwwCgYDVQQLDANPcmcxDzANBgNVBAMMBlJvb3RDQTCBnzANBgkqhkiG9w0B
AQEFAAOBjQAwgYkCgYEAoY0yhH5lba0xHg/2Ie51F4aFbU2LbBI3CyJGRry0RSCG
8rnb4WU3kcbHE03JEPQmihU1op73QmxI413mqPNtpqWYYOvL6B9gvVl0bEM5qTnD
hKfP9h/ekCO4pm+uY3JjiatkIF/W8/0cytbWsUgjfZU96S/uF4g48TxLe1H/Gt8C
AwEAAaNTMFEwHQYDVR0OBBYEFNqZ+NzXj2tP2NSwjk2HqInVvK56MB8GA1UdIwQY
MBaAFNqZ+NzXj2tP2NSwjk2HqInVvK56MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADgYEAe821E08aGjkGYEOK9bFRdzBgHgiA8BHupuz8fJ915YJJijwz
5DuVRfQSNGNrip5lZMImOv51hmeIyx7MjRyMfX5OTYl3N9jbpG9fN43apmc5LJSl
VH0XG4zByMNuL8siY1YWrbQ4M35K20i7h51NweWBqhC1tB0wd+dWShrZIJk=
-----END CERTIFICATE-----`
	pemServerCsr = `-----BEGIN CERTIFICATE REQUEST-----
MIIBdzCB4QIBADA4MQ0wCwYDVQQKDAR0ZXN0MQwwCgYDVQQLDANPcmcxGTAXBgNV
BAMMEGhvc3QuZXhhbXBsZS5jb20wgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGB
AMBUPG+Ws5Klgnt/XZeLkMcO0B2X0nF0bWMFbUjISM4+tM9eYgel3DY4nfbTpEJV
nrArRwQ/t+nxRd9Y9yvEo6MgEmWQBGr5LTZS2NYsCDxf+RxzsjzGQL1J14VFA2lL
xFvAOvoSgaImuLOWJiX9xgvPTw39PcSFm0/JRAEVP1flAgMBAAGgADANBgkqhkiG
9w0BAQsFAAOBgQBlr4bgtCa+crQ7MzkuhJQZjHnV3FkI88ZZo2/V1nhpMOw5EPDS
qM2Ajc3/eUqoj+X9z7iw05Yu5J4wmJvHUwp0OjVrPcE3CIiArSzL+s2HfjMHM0bF
zvRehNf9rzAatHoAQ0hmpppGB24NebQLl2qeDLVkRwtWoaxS3vKrfa+Fhw==
-----END CERTIFICATE REQUEST-----`
)

func TestCertInspectConfigMap(t *testing.T) {
	tests := []struct {
		name             string
		cm               *unstructured.Unstructured
		want             int
		output           string
		listNonCerts     bool
		showParseFailure bool
	}{
		{
			name:             "ConfigMap with CA certificate",
			cm:               getUnstructured("my-configmap", "my-namespace", "ConfigMap", map[string]string{"ca-bundle.crt": pemCaCert}),
			want:             1,
			output:           "",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "ConfigMap with invalid CA certificate",
			cm:               getUnstructured("my-configmap", "my-namespace", "ConfigMap", map[string]string{"ca-bundle.crt": "invalid ca"}),
			want:             0,
			output:           "data does not contain any valid RSA or ECDSA certificates",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "ConfigMap with empty CA certificate",
			cm:               getUnstructured("my-configmap", "my-namespace", "ConfigMap", map[string]string{"ca-bundle.crt": ""}),
			want:             0,
			output:           "missing content for key",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "ConfigMap without valid ca key ",
			cm:               getUnstructured("my-configmap", "my-namespace", "ConfigMap", map[string]string{"invalid-key": ""}),
			want:             0,
			output:           "does not contain a key",
			listNonCerts:     false,
			showParseFailure: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			listNonCerts = tt.listNonCerts
			showParseFailure = tt.showParseFailure
			res := inspectConfigMap(&output, tt.cm)
			fmt.Printf("%#v", res)
			if !strings.Contains(output.String(), tt.output) {
				t.Errorf("Got: %v \n", output.String())
				t.Errorf("Want: %v \n", tt.output)
			}
			if len(res) != tt.want {
				t.Errorf("Got: %d certificates\n", len(res))
				t.Errorf("Want: %d certificates\n", tt.want)
			}
		})
	}
}

func TestCertInspectSecret(t *testing.T) {
	tests := []struct {
		name             string
		secret           *unstructured.Unstructured
		want             int
		output           string
		listNonCerts     bool
		showParseFailure bool
	}{
		{
			name:             "Secret with server certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"tls.crt": string(encodeBase64(pemServerCert))}),
			want:             1,
			output:           "",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret with invalid server certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"tls.crt": string(encodeBase64("invalid cert"))}),
			want:             0,
			output:           "data does not contain any valid RSA or ECDSA certificates",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret with empty server certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"tls.crt": ""}),
			want:             0,
			output:           "missing content for key",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret with CA certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"ca.crt": string(encodeBase64(pemCaCert))}),
			want:             1,
			output:           "",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret with invalid CA certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"ca.crt": string(encodeBase64("invalid ca"))}),
			want:             0,
			output:           "data does not contain any valid RSA or ECDSA certificates",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret with empty CA certificate",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"ca.crt": ""}),
			want:             0,
			output:           "missing content for key",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "Secret without tls.crt or ca.crt key",
			secret:           getUnstructured("my-secret", "my-namespace", "Secret", map[string]string{"invalid key": ""}),
			want:             0,
			output:           "does not contain a key",
			listNonCerts:     false,
			showParseFailure: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			listNonCerts = tt.listNonCerts
			showParseFailure = tt.showParseFailure
			res := inspectSecret(&output, tt.secret)
			if !strings.Contains(output.String(), tt.output) {
				t.Errorf("Got: %v \n", output.String())
				t.Errorf("Want: %v \n", tt.output)
			}
			if len(res) != tt.want {
				t.Errorf("Got: %d certificates\n", len(res))
				t.Errorf("Want: %d certificates\n", tt.want)
			}
		})
	}
}

func TestCertInspectCSR(t *testing.T) {
	tests := []struct {
		name             string
		csr              *unstructured.Unstructured
		want             int
		output           string
		listNonCerts     bool
		showParseFailure bool
	}{
		{
			name:             "CertificateSigningRequest with valid certificate",
			csr:              getCertificateSigningRequest("my-csr", "my-namespace", []byte(pemServerCsr), []byte(pemServerCert)),
			want:             1,
			output:           "",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "CertificateSigningRequest with invalid certificate",
			csr:              getCertificateSigningRequest("my-csr", "my-namespace", []byte(pemServerCsr), encodeBase64("invalid cert")),
			want:             0,
			output:           "data does not contain any valid RSA or ECDSA certificates",
			listNonCerts:     false,
			showParseFailure: true,
		},
		{
			name:             "CertificateSigningRequest with invalid request",
			csr:              getCertificateSigningRequest("my-csr", "my-namespace", []byte("invalid"), encodeBase64("invalid")),
			want:             0,
			output:           "data does not contain any valid RSA or ECDSA certificates",
			listNonCerts:     false,
			showParseFailure: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			listNonCerts = tt.listNonCerts
			showParseFailure = tt.showParseFailure
			res := inspectCSR(&output, tt.csr)
			fmt.Printf("%#v", res)
			if !strings.Contains(output.String(), tt.output) {
				t.Errorf("Got: %v \n", output.String())
				t.Errorf("Want: %v \n", tt.output)
			}
			if len(res) != tt.want {
				t.Errorf("Got: %d certificates\n", len(res))
				t.Errorf("Want: %d certificates\n", tt.want)
			}
		})
	}
}

func getUnstructured(name, namespace, kind string, data map[string]string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"namespace":         name,
				"name":              namespace,
			},
			"data": data,
		},
	}
}

func getCertificateSigningRequest(name, namespace string, data, status []byte) *unstructured.Unstructured {
	obj := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request: data,
		},
		Status: certificatesv1.CertificateSigningRequestStatus{
			Certificate: status,
		},
	}
	unstructuredObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return &unstructured.Unstructured{Object: unstructuredObj}
}

func parsedCertificate(in []byte) *x509.Certificate {
	c, _ := x509.ParseCertificate(in)
	return c
}

func encodeBase64(data string) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, []byte(data))
	return dst
}
