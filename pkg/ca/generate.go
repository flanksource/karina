package ca

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/certs"
	"github.com/pkg/errors"
)

// GenerateCA generates a new CA certificate
func GenerateCA(name, certPath, privateKeyPath, password string, expiryYears int) error {
	if err := ensureDir(certPath, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directories for certificate path: %s", certPath)
	}
	if err := ensureDir(privateKeyPath, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directories for certificate private key path: %s", privateKeyPath)
	}

	ca := certs.NewCertificateBuilder(name).CA().Certificate
	signedCA, err := ca.SignCertificate(ca, expiryYears)
	if err != nil {
		return errors.Wrap(err, "failed to sign certificate")
	}

	if err := ioutil.WriteFile(certPath, signedCA.EncodedCertificate(), 0600); err != nil {
		return errors.Wrap(err, "failed to write certificate file")
	}

	encryptedPrivateKey := signedCA.EncodedPrivateKey()

	if password != "" {
		pk := x509.MarshalPKCS1PrivateKey(ca.PrivateKey)
		block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", pk, []byte(password), x509.PEMCipherAES256) // nolint: staticcheck
		if err != nil {
			return errors.Wrap(err, "failed to encrypt private key")
		}
		encryptedPrivateKey = pem.EncodeToMemory(block)
	}

	if err := ioutil.WriteFile(privateKeyPath, encryptedPrivateKey, 0600); err != nil {
		return errors.Wrap(err, "failed to write private key file")
	}
	return nil
}

func ensureDir(fileName string, mode os.FileMode) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, mode)
		if merr != nil {
			return merr
		}
	}
	return nil
}
