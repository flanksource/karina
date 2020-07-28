package ca

import (
	"fmt"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/types"
)

// ReadCA opens the CA stored in the file ca.Cert using the private key in ca.PrivateKey
// with key password ca.Password.
func ReadCA(ca *types.CA) (*certs.Certificate, error) {
	cert := files.SafeRead(ca.Cert)
	if cert == "" {
		return nil, fmt.Errorf("unable to read certificate %s", ca.Cert)
	}
	privateKey := files.SafeRead(ca.PrivateKey)
	if privateKey == "" {
		return nil, fmt.Errorf("unable to read private key %s", ca.PrivateKey)
	}
	return certs.DecryptCertificate([]byte(cert), []byte(privateKey), []byte(ca.Password))
}
