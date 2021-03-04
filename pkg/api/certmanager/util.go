package certmanager

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const APIVersion = "cert-manager.io/v1"
const DefaultIsser = "default-issuer"

func NewCertificateForService(namespace string, name string) Certificate {
	return Certificate{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       CertificateKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: CertificateSpec{
			DNSNames: []string{
				name,
				fmt.Sprintf("%s.%s.svc", name, namespace),
				fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			},
			SecretName: name,
			IssuerRef: ObjectReference{
				Kind: "ClusterIssuer",
				Name: DefaultIsser,
			},
		},
	}
}
