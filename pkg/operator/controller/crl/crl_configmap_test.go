package crl

import (
	"crypto/x509/pkix"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// exampleRootCA is a PEM-encoded self-signed CA certificate generated
	// using the following commands:
	//
	//     openssl genrsa -out root-ca.key 2048
	//     openssl req -new -key root-ca.key -out root-ca.csr -subj "/C=US/ST=NC/L=Chocowinity/O=OS3/OU=Eng/CN=Test Root CA"
	//     openssl x509 -req -in root-ca.csr -out root-ca.crt -days 3650 -signkey root-ca.key -extfile root-ca.cnf
	//
	// where the content of root-ca.cnf is the following two lines
	// (excluding indentation):
	//
	//     crlDistributionPoints=URI:http://example/root.crl
	//     subjectKeyIdentifier=hash
	exampleRootCA = `-----BEGIN CERTIFICATE-----
MIIDnzCCAoegAwIBAgIUSij3jKUFtUiGbsOs556UepNPtyAwDQYJKoZIhvcNAQEL
BQAwYzELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAk5DMRQwEgYDVQQHDAtDaG9jb3dp
bml0eTEMMAoGA1UECgwDT1MzMQwwCgYDVQQLDANFbmcxFTATBgNVBAMMDFRlc3Qg
cm9vdCBDQTAeFw0yMjA4MjcwMjUwMDhaFw0yMjA4MjgwMjUwMDhaMGMxCzAJBgNV
BAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkxDDAKBgNV
BAoMA09TMzEMMAoGA1UECwwDRW5nMRUwEwYDVQQDDAxUZXN0IHJvb3QgQ0EwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCtqxuaxQYpaDFY/WYeavNpKdPe
Q3zfYqpmoQLo9yWhm4L8pssGoapz6z6te1ZVq7e/UgZZ87QQzMLM5HEjtMmmXnrg
3cQD7O0/PEjgd5VHyynGvfqJipgM+KL6GdzFkDHiSeDDbTlsHftnHdh6/Q2T+Kjk
piSxFp2xOBHsZHPH4Ecr2NBvxfCAzZ8hVO20ivApO6BMyL+RazaUBhGKCelXkfZL
2E3V+v50HcyY5yB4DC+rMFlagns17gfl8qcx9Fvg+W5v5WiTfGfe3QjlxQ1Yolgg
OdHUjYrDlMxgDoN4jaDUrOFi9ERqIN796obDrOQR2+XpF6W1h0AQR0y2NHNTAgMB
AAGjSzBJMCgGA1UdHwQhMB8wHaAboBmGF2h0dHA6Ly9leGFtcGxlL3Jvb3QuY3Js
MB0GA1UdDgQWBBTnXcaWUgNTaD+MpEYgTACc9WUOZTANBgkqhkiG9w0BAQsFAAOC
AQEANbqseu1kcZ+rUDb4j1Ti/T2fJNeJnlPQDsJ4w3hrp+MimFG/q1w7t2f7xa6F
bgPkOf6w4VHz7rRqMp5VNNDjYP9KDdxsFKXfVi7WRhEbSmiamwQgasH1J0K+bZuU
5ucq3pscu8hdYErwW7dosfhvijg+ykrsec9IYprtvWYs/4UZKm54i8/Q4yQO9tOg
F30kXRDdfIBLTUgjKYXXNNOGFDy3SqJFPGKauB5KCZ7CEx4E2p1x+8uwL22XH26e
HZJcfSAJZkEc15KoLukyQlwzXeVGmhPY7rUC/fOf8E1pDtUf5q3Rtmz1Z2epBZdk
T3YH5JqMZeAL/+Z4PACTSNFTGA==
-----END CERTIFICATE-----
`

	// exampleIntermediateCA is a PEM-encoded self-signed CA certificate
	// generated using the following commands:
	//
	//     openssl genrsa -out intermediate-ca.key 2048
	//     openssl req -new -key intermediate-ca.key -out intermediate-ca.csr -subj "/C=US/ST=NC/L=Chocowinity/O=OS3/OU=Eng/CN=Test Intermediate CA"
	//     openssl x509 -req -in intermediate-ca.csr -out intermediate-ca.crt -days 3650 -CA root-ca.crt -CAcreateserial -CAkey root-ca.key -extfile intermediate-ca.cnf
	//
	// where the content of intermediate-ca.cnf is the following three lines
	// (excluding indentation):
	//
	//     crlDistributionPoints=URI:http://example/intermediate.crl
	//     subjectKeyIdentifier=hash
	//     authorityKeyIdentifier=keyid,issuer
	exampleIntermediateCA = `-----BEGIN CERTIFICATE-----
MIID0DCCArigAwIBAgIUMCDJzIpGFRrrBu2Ne4rR3IEEbCYwDQYJKoZIhvcNAQEL
BQAwYzELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAk5DMRQwEgYDVQQHDAtDaG9jb3dp
bml0eTEMMAoGA1UECgwDT1MzMQwwCgYDVQQLDANFbmcxFTATBgNVBAMMDFRlc3Qg
cm9vdCBDQTAeFw0yMjA4MjcwMjUwMTZaFw0zMjA4MjQwMjUwMTZaMGsxCzAJBgNV
BAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkxDDAKBgNV
BAoMA09TMzEMMAoGA1UECwwDRW5nMR0wGwYDVQQDDBRUZXN0IGludGVybWVkaWF0
ZSBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANk6R5pqCmLI59lf
bAJgqzTOXXnHBCHee5pM71geZ5/O7i0n1wXvvZ77vvMQri++w8Y/9kyQ8mKWDvRS
nFmeiFupTix/vsm0MaWNxoVYkkgnc3A/EGG56Z4hzTdYM0koIrqqaAEh6gkIBNew
iI4TUeCnHS0m8Eh7ayd2/Uuz+Vn+jKqlwiz9tpvMhQjvkT0O3QK6WwAX570ovcWs
QWrA29PUhADgh/P0Inab6mD7Wzx5B1Gew//X+Og80SJAEVZK76vr2sEDBT7rhxp4
CbpgG6Yee7rl8hkG5DYAjwnBud9TFz/SWlGf13m/ygQZs7V4RVuXu1B2hwfOuql+
MWPNSZsCAwEAAaN0MHIwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL2V4YW1wbGUv
aW50ZXJtZWRpYXRlLmNybDAdBgNVHQ4EFgQUaiNY4rih6m+px7pEuqVof5uun2sw
HwYDVR0jBBgwFoAU513GllIDU2g/jKRGIEwAnPVlDmUwDQYJKoZIhvcNAQELBQAD
ggEBAKk9YLt4OHv85SjKsQuuTOOXa7mMFYH9tMEXBbxk54BfEFzUnXf0Dofc6r3E
ZC5tkdw7QWNQ94VcHh+hlD+HL7Ceidr7ZDPfXxPVNmbuz/0HZJU8k7NzzkQvs3dg
3hfvlTqbWQzaspvi5ZCUq3K/4DmeipVHKb2vvvZplGnSTTt1emtO3YR1Hx716MP7
BnYYE2+CQTizWEJ6SWHTPfZzIK04VtpeCZqW1bbbOQVkn0VS4FQzycxvjrD4PoH2
KsPpVupKaBzF/rrHt/GfgG63bsLVmzYHQZcAWhMYczvfFB56f5eGkWkfhn+nprPJ
lloWAn6P0sWqUaIg3C8Cs6l1o/M=
-----END CERTIFICATE-----
`

	// exampleClientCertificate is a PEM-encoded client certificate that is
	// signed by the CA in exampleIntermediateCA.  The client certificate
	// was generated using the following OpenSSL commands:
	//
	//     openssl genrsa -out client1.key 2048
	//     openssl req -new -key client1.key -out client1.csr -subj "/C=US/ST=NC/L=Chocowinity/O=OS3/OU=Eng/CN=Test Client"
	//     openssl x509 -req -days 1 -in client1.csr -signkey client1.key -CA intermediate-ca.crt -CAcreateserial -CAkey intermediate-ca.key -out client1.crt
	exampleClientCertificate = `-----BEGIN CERTIFICATE-----
MIIDVDCCAjwCFE8B3wSROV+V4MLsI/pTvk4bTnBuMA0GCSqGSIb3DQEBCwUAMGsx
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkx
DDAKBgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMR0wGwYDVQQDDBRUZXN0IGludGVy
bWVkaWF0ZSBDQTAeFw0yMjA4MjcwMjUwMjdaFw0yMjA4MjgwMjUwMjdaMGIxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkxDDAK
BgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMRQwEgYDVQQDDAtUZXN0IENsaWVudDCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKKHCEu5JYxX6DPNPrK+VMEH
E5/Q7iQFxGV8QF+whLVB9xB0LIl8bRIwqJcQ+lXs4e3HaZb8X6eaDvDk64zBnL5O
HGkcX8OMrVKHHeetc1yle7SO92f9vaYlGhsTPs4/AyELAhDWAS+Oc1iDbTADHyoI
vDgdABwz0ZJAzO2xkijztS2y3crRN3vl38Oqm2ysVF+uAeg6eFMxLrSyGS//1X7U
39lYQr4RE7hwxf8Siol1s6hNcngcU4XFI6yWipkUaJ0VjH5seDtANUwNhUkRVxvt
TSUk2czW57leP1QNT/M2sYTZE2+E2hPoHoQS2ks3XJ9ncraW4GdWrvJ0vFztp9cC
AwEAATANBgkqhkiG9w0BAQsFAAOCAQEAN89qWI+ofiZKCbnn7Cyx2I5A/K64tb8B
uK5jP1uqdxg36ftTiTNlGj8We5LMuXfQGuMNC/BuccE4lyHjh+yKg3LxoR8KIDGp
Ria6MVvrE2YFVCElh1OqYEAGmnr9MoCzKIPjyuAmI+5QD202zHHwKPC4HJ+Tt0XZ
B4bpbyxPCEZFwniJSxxUWx2u7wVozw7z8PghwQrJsQHt9hsVLbF1t/csSgH0ThD4
PqxnQPpC5nlZz/3r328deJVYF9IG8tmAwnO4kCmpla2Jjnm+VL60dDT7QliwZzE+
WpqbLiboLOxTfl3FXEm4hMIFmMARQ2mosgHbWGW28Ttxe17Z4wuK2Q==
-----END CERTIFICATE-----
`

	// exampleRevokedClientCertificate is a PEM-encoded client certificate
	// that is signed by the CA in exampleIntermediateCA.  The client
	// certificate was generated using the following OpenSSL commands:
	//
	//     openssl genrsa -out client2.key 2048
	//     openssl req -new -key client2.key -out client2.csr -subj "/C=US/ST=NC/L=Chocowinity/O=OS3/OU=Eng/CN=Another Test Client"
	//     openssl x509 -req -days 1 -in client2.csr -signkey client2.key -CA intermediate-ca.crt -CAcreateserial -CAkey intermediate-ca.key -out client2.crt
	exampleRevokedClientCertificate = `-----BEGIN CERTIFICATE-----
MIIDXDCCAkQCFE8B3wSROV+V4MLsI/pTvk4bTnBvMA0GCSqGSIb3DQEBCwUAMGsx
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkx
DDAKBgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMR0wGwYDVQQDDBRUZXN0IGludGVy
bWVkaWF0ZSBDQTAeFw0yMjA4MjcwMjUwMzBaFw0yMjA4MjgwMjUwMzBaMGoxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkxDDAK
BgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMRwwGgYDVQQDDBNBbm90aGVyIFRlc3Qg
Q2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAza6/zS7ejg8X
1+h0Ag1u95nr1D4ZrfQJnqHsQJvXwDd+UlAQW1RrWpKmg5Kw3AALa5caouSGg8iB
leeTGkAIe8AgUrPL00O8/0XeOxzLAUQmw2IEWppYt6oOAf2BeTCMC6dHmWSNdzsk
QMijosAz4f9qIJNprhAWO9bQTROv2ldH5lowHOqKq72Jp31/ZFhdd/gZLPbKY97F
/BGsJbi+A9w7n988ZPS6K0AhioJMWH2RT+xwcbsbPHgwT+gc7cAIPptAO8JsGsZH
aYltXvtQazBkvtoECi/y5DckEZ55G0PgGYYeFXu7dvuq8qtmvEWqTB5EnBi39ipw
NqpDXFozYQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQCyKFIACevpsuDtW425ihDW
69S3JrlLHk1XUfv49H54x2CClsPH3qg7wu6cl4FAXrZxZd4/OCuTyh/DHx5JWvAf
bJ4iSxIkQGb1lOSuHVM2iYJ6mu7BT/aUy6BW0SYrlws6u2gSOFOekG7WzdSM4nLP
l5bCx91YzRCAf4AHHTGEA3hVHOxsBZR0vYLGiBRxu+0zPngf7AaPLNCpaH0w6NTo
egjQGm9lhxl0jU0PKZXZBcXwsVRGkF4bd6Pu/dYjbfCdeBgCoIpqKTlgDMlrGOV+
Ij4IPxfKm7nUr7iu3uNc358xlfAVJJVnMo8ZTRnH5eWECSD5TtzD0kG07392jLAB
-----END CERTIFICATE-----
`

	// exampleClientCertWithCRL is a PEM-encoded client certificate that is
	// signed by the CA in exampleIntermediateCA and specifies an additional
	// CRL distribution point.  The client certificate was generated using
	// the following OpenSSL commands:
	//
	//     openssl genrsa -out client3.key 2048
	//     openssl req -new -key client3.key -out client3.csr -subj "/C=US/ST=NC/L=Chocowinity/O=OS3/OU=Eng/CN=Test Client #3"
	//     openssl x509 -req -days 1 -in client3.csr -signkey client3.key -CA intermediate-ca.crt -CAcreateserial -CAkey intermediate-ca.key -out client3.crt
	exampleClientCertWithCRL = `-----BEGIN CERTIFICATE-----
MIIDVDCCAjwCFE8B3wSROV+V4MLsI/pTvk4bTnBuMA0GCSqGSIb3DQEBCwUAMGsx
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkx
DDAKBgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMR0wGwYDVQQDDBRUZXN0IGludGVy
bWVkaWF0ZSBDQTAeFw0yMjA4MjcwMjUwMjdaFw0yMjA4MjgwMjUwMjdaMGIxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJOQzEUMBIGA1UEBwwLQ2hvY293aW5pdHkxDDAK
BgNVBAoMA09TMzEMMAoGA1UECwwDRW5nMRQwEgYDVQQDDAtUZXN0IENsaWVudDCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKKHCEu5JYxX6DPNPrK+VMEH
E5/Q7iQFxGV8QF+whLVB9xB0LIl8bRIwqJcQ+lXs4e3HaZb8X6eaDvDk64zBnL5O
HGkcX8OMrVKHHeetc1yle7SO92f9vaYlGhsTPs4/AyELAhDWAS+Oc1iDbTADHyoI
vDgdABwz0ZJAzO2xkijztS2y3crRN3vl38Oqm2ysVF+uAeg6eFMxLrSyGS//1X7U
39lYQr4RE7hwxf8Siol1s6hNcngcU4XFI6yWipkUaJ0VjH5seDtANUwNhUkRVxvt
TSUk2czW57leP1QNT/M2sYTZE2+E2hPoHoQS2ks3XJ9ncraW4GdWrvJ0vFztp9cC
AwEAATANBgkqhkiG9w0BAQsFAAOCAQEAN89qWI+ofiZKCbnn7Cyx2I5A/K64tb8B
uK5jP1uqdxg36ftTiTNlGj8We5LMuXfQGuMNC/BuccE4lyHjh+yKg3LxoR8KIDGp
Ria6MVvrE2YFVCElh1OqYEAGmnr9MoCzKIPjyuAmI+5QD202zHHwKPC4HJ+Tt0XZ
B4bpbyxPCEZFwniJSxxUWx2u7wVozw7z8PghwQrJsQHt9hsVLbF1t/csSgH0ThD4
PqxnQPpC5nlZz/3r328deJVYF9IG8tmAwnO4kCmpla2Jjnm+VL60dDT7QliwZzE+
WpqbLiboLOxTfl3FXEm4hMIFmMARQ2mosgHbWGW28Ttxe17Z4wuK2Q==
-----END CERTIFICATE-----
`

	// exampleRootCACRL is a PEM-encoded revocation certificate list for the
	// CA in exampleRootCA.  This list is empty.  exampleRootCACRL was
	// generated using the following commands:
	//
	//     touch index.txt
	//     openssl ca -gencrl -out root-ca.crl -config root-ca.cnf 
	exampleRootCACRL = `-----BEGIN X509 CRL-----
MIIBqTCBkjANBgkqhkiG9w0BAQsFADBjMQswCQYDVQQGEwJVUzELMAkGA1UECAwC
TkMxFDASBgNVBAcMC0Nob2Nvd2luaXR5MQwwCgYDVQQKDANPUzMxDDAKBgNVBAsM
A0VuZzEVMBMGA1UEAwwMVGVzdCByb290IENBFw0yMjA4MjcwMjUxMDNaFw0yMjA4
MjcwMzUxMDNaMA0GCSqGSIb3DQEBCwUAA4IBAQAWmlpItXR3TkDa5+oJmXqoeOb0
2R1b5E5UJFCcVxKJSUdSoyLwPdRcE6dExxlW4TH+oSOa7m5RjP/sFsNw/prlchxn
XPDfPWm3QVfDpWUvRsVV0yeseY/gqQtM7ASLhjvqLeZBwPjKREyohiU9zB80jt5N
NiGw+2DLOxswNLIM6JrcmIiyimfncR/pibvPiqgzcvcp+m9reL1dIF3/7I3/IB3r
X4ehhYyaGv5MY5MALN4Q1VtbiJ450Dkbx9VRi7JU606lFFd4vqlr244cqwJMi5n0
zC3V3pCM6v15mMBL2qiRQnJFA5ukwWQKWzcm5IIY+KntbErtOfjcGgFZo+IM
-----END X509 CRL-----
`

	// exampleIntermediateCACRL is a PEM-encoded revocation certificate list
	// for the CA in exampleIntermediateCA.  This list includes the client
	// certificate in exampleRevokedClientCertificate.
	// exampleIntermediateCACRL was generated using the following commands:
	//
	//     openssl ca -revoke client2.crt -config intermediate-ca.cnf
	//     openssl ca -gencrl -out intermediate-ca.crl -config intermediate-ca.cnf 
	exampleIntermediateCACRL = `-----BEGIN X509 CRL-----
MIIB2jCBwzANBgkqhkiG9w0BAQsFADBrMQswCQYDVQQGEwJVUzELMAkGA1UECAwC
TkMxFDASBgNVBAcMC0Nob2Nvd2luaXR5MQwwCgYDVQQKDANPUzMxDDAKBgNVBAsM
A0VuZzEdMBsGA1UEAwwUVGVzdCBpbnRlcm1lZGlhdGUgQ0EXDTIyMDgyNzAyNTEy
NloXDTIyMDgyNzAzNTEyNlowJzAlAhRPAd8EkTlfleDC7CP6U75OG05wbxcNMjIw
ODI3MDI1MTI2WjANBgkqhkiG9w0BAQsFAAOCAQEAsOPGM0nIsWCRHhOnO+F4zEyn
OrIEVC1/Ckn8J3tO5EsZlW55U176LdbJY7q96tkDX9XW1BY+4FtpB3+r+ZqEtJfJ
nwM2pqbBLczZqo/S77O7iLNf5nbinwCX+aYoysUwy/I24IHHlu0LBTy4pXGnKszf
wDPYoyjLNyER6jIe7IobQ4MUow1Y9AV2vE+hZv7yIVLE/ZDqQyCZdK3VwpuolvXB
5L7XNQ1a0WiGI7BiTUfKuS7WRzyD5RLFsjiL3ZGT8Z2T29rS8x5KKWUENcXLMdpG
ZK4ig9l6s0r/3gzRgyi5yi6GjutpkBdvgXIcP67fyg581p9YbOYZTZnyMmmWYw==
-----END X509 CRL-----
`
)

// // formatCRL formats a map[string]*pkix.CertificateList value for use in test output.
// func formatCRL(crl map[string]*pkix.CertificateList) string {
// 	buf := &bytes.Buffer{}
// 	fmt.Fprint(buf, "map[string]*pkix.CertificateList{\n")
// 	for k, v := range crl {
// 		fmt.Fprintf(buf, "    %q: &pkix.CertificateList{\"\n", k)
// 		fmt.
// 	}
// }

func Test_buildCRLMap(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedOutput map[string]*pkix.CertificateList
		expectError    bool
	}{{
		name:        "empty CRL",
		input:       []byte{},
		expectedOutput: map[string]*pkix.CertificateList{},
		expectError: false,
	}, {
		name:        "invalid input",
		input:       []byte("nonsense"),
		expectedOutput: map[string]*pkix.CertificateList{},
		expectError: false,
	}, {
		name:           "valid CRL",
		input:          []byte(exampleIntermediateCACRL),
		expectedOutput: map[string]*pkix.CertificateList{
			// XXX
		},
		expectError:    false,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//buildCRLMap(crlData []byte) (map[string]*pkix.CertificateList, error)
			actualCRL, err := buildCRLMap(tc.input)
			switch {
			case err == nil && tc.expectError:
				t.Fatal("expected error, got nil")
			case err != nil && !tc.expectError:
				t.Fatalf("unexpected error: %v", err)
			case !reflect.DeepEqual(actualCRL, tc.expectedOutput):
				t.Fatalf("expected:\n%v\ngot:\n%v", spew.Sdump(tc.expectedOutput), spew.Sdump(actualCRL))
			}
		})
	}
}

func Test_desiredCRLConfigMap(t *testing.T) {
	makeIC := func(icName, policy, caName string) *operatorv1.IngressController {
		return &operatorv1.IngressController{
			ObjectMeta: metav1.ObjectMeta{
				Name:      icName,
				Namespace: "openshift-ingress-operator",
			},
			Spec: operatorv1.IngressControllerSpec{
				ClientTLS: operatorv1.ClientTLS{
					ClientCertificatePolicy: operatorv1.ClientCertificatePolicy(policy),
					ClientCA: configv1.ConfigMapNameReference{
						Name: caName,
					},
				},
			},
		}
	}
	testCases := []struct {
		name              string
		ic                *operatorv1.IngressController
		clientCAData      []byte
		crls              map[string]*pkix.CertificateList
		expectedConfigMap *corev1.ConfigMap
		expectError       bool
	}{{
		name:              "empty policy",
		ic:                makeIC("custom", "", "client-ca"),
		expectedConfigMap: nil,
		expectError:       false,
	}, {
		name:              "empty certificate reference",
		ic:                makeIC("custom", "Optional", ""),
		expectedConfigMap: nil,
		expectError:       false,
	}, {
		name:         "intermediate CA, no previously downloaded CRL",
		ic:           makeIC("custom", "Required", "client-ca"),
		clientCAData: []byte(exampleIntermediateCA),
		crls:         map[string]*pkix.CertificateList{},
		expectedConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"crl.pem": exampleIntermediateCACRL,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "router-client-ca-crl-custom",
				Namespace: "openshift-ingress",
			},
		},
		expectError:  false,
	}, {
		name:              "root CA + intermediate CA, no previously downloaded CRL",
		ic:                makeIC("custom", "Required", "client-ca"),
		clientCAData:      []byte(exampleRootCA + exampleIntermediateCA),
		crls:              map[string]*pkix.CertificateList{},
		expectedConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"crl.pem": exampleRootCACRL + exampleIntermediateCACRL,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "router-client-ca-crl-custom",
				Namespace: "openshift-ingress",
			},
		},
		expectError: false,
	}, {
		name:         "root CA + intermediate CA, CRL already up to date",
		ic:                makeIC("custom", "Required", "client-ca"),
		clientCAData: []byte(exampleRootCA + exampleIntermediateCA),
		crls:         map[string]*pkix.CertificateList{
			// XXX
		},
		expectedConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"crl.pem": exampleRootCACRL + exampleIntermediateCACRL,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "router-client-ca-crl-custom",
				Namespace: "openshift-ingress",
			},
		},
		expectError: false,
	}, {
		name:         "intermediate CA + client cert with a distribution point, CRL already up to date",
		ic:                makeIC("custom", "Required", "client-ca"),
		clientCAData: []byte(exampleIntermediateCA + exampleClientCertWithCRL),
		crls:         map[string]*pkix.CertificateList{
			// XXX
		},
		expectedConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"crl.pem": exampleIntermediateCACRL + "" /*exampleExtraCRL*/,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "router-client-ca-crl-custom",
				Namespace: "openshift-ingress",
			},
		},
		expectError: false,
	}}

	httpGet = func(url string) (*http.Response, error) {
		var (
			err error
			body []byte
		)
		switch url {
		case "http://example/root.crl":
			body = []byte(exampleRootCACRL)
		case "http://example/intermediate.crl":
			body = []byte(exampleIntermediateCACRL)
		default:
			err = fmt.Errorf("fake error for url %s", url)
		}
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader(body))}, err
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//desiredCRLConfigMap(ic *operatorv1.IngressController, ownerRef metav1.OwnerReference, clientCAData []byte, crls map[string]*pkix.CertificateList) (bool, *corev1.ConfigMap, error)
			trueVar := true
			ownerRef := metav1.OwnerReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "router-custom",
				UID:        "1",
				Controller: &trueVar,
			}
			var data []byte
			if len(tc.clientCAData) != 0 {
				data = make([]byte, len(tc.clientCAData))
				copy(data, tc.clientCAData)
			}
			haveCM, actualCM, err := desiredCRLConfigMap(tc.ic, ownerRef, tc.clientCAData, tc.crls)
			if !reflect.DeepEqual(data, tc.clientCAData) {
				t.Fatalf("desiredCRLConfigMap mutated clientCAData!\nold: %s\nnew: %s", data, tc.clientCAData)
			}
			if haveCM != (actualCM != nil) {
				t.Fatalf("desiredCRLConfigMap returned %v, %v, %v", haveCM, actualCM, err)
			}
			if tc.expectedConfigMap != nil {
				tc.expectedConfigMap.ObjectMeta.OwnerReferences = []metav1.OwnerReference{ownerRef}
			}
			switch {
			case err == nil && tc.expectError:
				t.Fatal("expected error, got nil")
			case err != nil && !tc.expectError:
				t.Fatalf("unexpected error: %v", err)
			case !reflect.DeepEqual(tc.expectedConfigMap, actualCM):
				t.Fatalf("expected:\n%v\ngot:\n%v", spew.Sdump(tc.expectedConfigMap), spew.Sdump(actualCM))
			}
			// t.Fatal("foo")
		})
	}
	httpGet = http.Get
}
