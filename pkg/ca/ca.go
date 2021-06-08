package ca

import (
	"fmt"
	"strings"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/is"
	"github.com/flanksource/karina/pkg/types"
)

// ReadCA opens the CA stored in the file ca.Cert using the private key in ca.PrivateKey
// with key password ca.Password.
func ReadCA(ca *types.CA) (*certs.Certificate, error) {
	var cert, privateKey string
	if strings.HasPrefix(ca.Cert, "-----BEGIN CERTIFICATE-----") {
		cert = ca.Cert
	} else {
		cert = files.SafeRead(ca.Cert)
	}

	if strings.HasPrefix(ca.PrivateKey, "-----BEGIN RSA PRIVATE KEY-----") {
		privateKey = ca.PrivateKey
	} else if is.File(ca.PrivateKey) {
		privateKey = files.SafeRead(ca.PrivateKey)
	} else {
		privateKey = ca.PrivateKey
	}

	if cert == "" {
		return nil, fmt.Errorf("unable to read certificate %s", ca.Cert)
	}

	if privateKey == "" {
		return nil, fmt.Errorf("unable to read private key %s", ca.PrivateKey)
	}

	if ca.Password == "" {
		return certs.DecodeCertificate([]byte(cert), []byte(privateKey))
	}
	return certs.DecryptCertificate([]byte(cert), []byte(privateKey), []byte(ca.Password))
}
