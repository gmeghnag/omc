package certs

import (
	"bytes"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/cert"
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
		name string
		cm   *corev1.ConfigMap
		want string
	}{
		{
			name: "ConfigMap with CA certificate",
			cm:   getConfigMap(map[string]string{"ca-bundle.crt": pemCaCert}),
			want: "\"RootCA\" [] groups=[test] issuer=\"<self>\"",
		},
		{
			name: "ConfigMap with invalid CA certificate",
			cm:   getConfigMap(map[string]string{"ca-bundle.crt": "invalid ca"}),
			want: "ERROR - data does not contain any valid RSA or ECDSA certificates\n",
		},
		{
			name: "ConfigMap with empty CA certificate",
			cm:   getConfigMap(map[string]string{"ca-bundle.crt": ""}),
			want: "MISSING ca-bundle content",
		},
		{
			name: "ConfigMap without ca-bundle.crt key ",
			cm:   getConfigMap(map[string]string{"invalid key": "invalid ca"}),
			want: "NOT a ca-bundle",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bytes.Buffer
			inspectConfigMap(&result, tt.cm)
			if !strings.Contains(result.String(), tt.want) {
				t.Errorf("Got: %v \n", result.String())
				t.Errorf("Want: %v \n", tt.want)
			}
		})
	}
}

func TestCertInspectSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret *corev1.Secret
		want   string
	}{
		{
			name:   "Secret with server certificate",
			secret: getSecret(map[string][]byte{"tls.crt": []byte(pemServerCert)}),
			want:   "\"system:nodes:mynode.example.com\" [serving] groups=[system:nodes] validServingFor=[system:nodes:mynode.example.com,192.168.1.1] issuer=\"kube-csr-signer_@123456789\"",
		},
		{
			name:   "Secret with invalid server certificate",
			secret: getSecret(map[string][]byte{"tls.crt": []byte("invalid pem")}),
			want:   "ERROR - data does not contain any valid RSA or ECDSA certificates\n",
		},
		{
			name:   "Secret with empty server certificate",
			secret: getSecret(map[string][]byte{"tls.crt": []byte("")}),
			want:   "MISSING tls.crt content",
		},
		{
			name:   "Secret with CA certificate",
			secret: getSecret(map[string][]byte{"ca.crt": []byte(pemCaCert)}),
			want:   "\"RootCA\" [] groups=[test] issuer=\"<self>\"",
		},
		{
			name:   "Secret with invalid CA certificate",
			secret: getSecret(map[string][]byte{"ca.crt": []byte("invalid ca")}),
			want:   "ERROR - data does not contain any valid RSA or ECDSA certificates\n",
		},
		{
			name:   "Secret with empty CA certificate",
			secret: getSecret(map[string][]byte{"ca.crt": []byte("")}),
			want:   "MISSING ca.crt content",
		},
		{
			name:   "Secret without tls.crt or ca.crt key",
			secret: getSecret(map[string][]byte{"invalid key": []byte("invalid pem")}),
			want:   "NOT a tls secret or token secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bytes.Buffer
			inspectSecret(&result, tt.secret)
			//			t.Errorf("Get: %s\nNeed:%s\n", result.String(), tt.want)
			if !strings.Contains(result.String(), tt.want) {
				t.Errorf("Got: %v \n", result.String())
				t.Errorf("Want: %v \n", tt.want)
			}
		})
	}
}

func TestCertificateSigningRequest(t *testing.T) {
	tests := []struct {
		name string
		csr  *certificatesv1.CertificateSigningRequest
		want string
	}{
		{
			name: "valid CertificateSigningRequest",
			csr:  getCertificateSigningRequest([]byte(pemServerCsr), []byte(pemServerCert)),
			want: "\"system:nodes:mynode.example.com\" [serving] groups=[system:nodes] validServingFor=[system:nodes:mynode.example.com,192.168.1.1] issuer=\"kube-csr-signer_@123456789\"",
		},
		{
			name: "unsigned CertificateSigningRequest",
			csr:  getCertificateSigningRequest([]byte(pemServerCsr), []byte("")),
			want: "NOT SIGNED",
		},
		{
			name: "invalid CertificateSigningRequest",
			csr:  getCertificateSigningRequest([]byte("invalid"), []byte("abc")),
			want: "ERROR -",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bytes.Buffer
			inspectCSR(&result, tt.csr)
			if !strings.Contains(result.String(), tt.want) {
				t.Errorf("Got: %v \n", result.String())
				t.Errorf("Want: %v \n", tt.want)
			}
		})
	}
}

func TestCertDetail(t *testing.T) {
	tests := []struct {
		name string
		cert string
		want string
	}{
		{
			name: "valid node Certificate",
			cert: pemServerCert,
			want: "\"system:nodes:mynode.example.com\" [serving] groups=[system:nodes] validServingFor=[system:nodes:mynode.example.com,192.168.1.1] issuer=\"kube-csr-signer_@123456789\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedCert, err := cert.ParseCertsPEM([]byte(tt.cert))
			if err != nil {
				t.Errorf("err: %s", err)
			}
			for _, cert := range parsedCert {
				result := certDetail(cert)
				if !strings.Contains(result, tt.want) {
					t.Errorf("Got: %v \n", result)
					t.Errorf("Want: %v \n", tt.want)
				}
			}
		})
	}
}

func getConfigMap(data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-configmap",
			Namespace: "my-namespace",
		},
		Data: data,
	}
}

func getSecret(data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "openshift-config",
		},
		Data: data,
	}
}

func getCertificateSigningRequest(data, status []byte) *certificatesv1.CertificateSigningRequest {
	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-csr",
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request: data,
		},
		Status: certificatesv1.CertificateSigningRequestStatus{
			Certificate: status,
		},
	}
}
