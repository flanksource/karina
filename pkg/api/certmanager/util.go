package certmanager

import (
	"fmt"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const APIVersion = "cert-manager.io/v1"
const DefaultIsser = "default-issuer"

func NewCertificateForService(namespace string, name string) certmanagerv1.Certificate {
	return certmanagerv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       certmanagerv1.CertificateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				certmanagerv1.AllowsInjectionFromSecretAnnotation: "true",
			},
		},
		Spec: certmanagerv1.CertificateSpec{
			DNSNames: []string{
				name,
				fmt.Sprintf("%s.%s.svc", name, namespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			},
			SecretName: name,
			IssuerRef: v1.ObjectReference{
				Kind: "ClusterIssuer",
				Name: DefaultIsser,
			},
			PrivateKey: &certmanagerv1.CertificatePrivateKey{
				Algorithm: certmanagerv1.RSAKeyAlgorithm,
				Size:      2048,
			},
		},
	}
}
