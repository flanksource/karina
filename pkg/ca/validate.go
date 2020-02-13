package ca

import (
	"fmt"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	"github.com/pkg/errors"
)

func ValidateCA(certPath, privateKeyPath, password string) error {
	certKey := files.SafeRead(certPath)
	privateKey := files.SafeRead(privateKeyPath)

	var cert *certs.Certificate
	var err error

	if password != "" {
		cert, err = certs.DecryptCertificate([]byte(certKey), []byte(privateKey), []byte(password))
		if err != nil {
			return errors.Wrap(err, "failed to decrypt certificate")
		}
	} else {
		cert, err = certs.DecodeCertificate([]byte(certKey), []byte(privateKey))
		if err != nil {
			return errors.Wrap(err, "failed to decrypt certificate")
		}
	}

	hash, err := cert.GetHash()
	if err != nil {
		return errors.Wrap(err, "failed to get certificate hash")
	}
	fmt.Printf("Certificate hash is %s\n", hash)

	fmt.Printf("Expires at: %s\n", cert.X509.NotAfter)

	return nil
}
