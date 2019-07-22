package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math"
	"math/big"
	"strings"

	"github.com/pkg/errors"
	"net"
	"time"
)

const (
	rsaKeySize   = 2048
	duration365d = time.Hour * 24 * 365
)

// NewPrivateKey creates an RSA private key
func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

// AltNames contains the domain names and IP addresses that will be added
// to the API Server's x509 certificate SubAltNames field. The values will
// be passed directly to the x509.Certificate object.
type AltNames struct {
	DNSNames []string
	IPs      []net.IP
}

// Config contains the basic fields required for creating a certificate
type Config struct {
	CommonName   string
	Organization []string
	AltNames     AltNames
	Usages       []x509.ExtKeyUsage
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func (cfg *Config) NewSignedCert(key *rsa.PrivateKey, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate random integer for signed cerficate")
	}

	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}

	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	tmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}

	b, err := x509.CreateCertificate(rand.Reader, &tmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create signed certificate: %+v", tmpl)
	}

	return x509.ParseCertificate(b)
}

// NewCertificateAuthority creates new certificate and private key for the certificate authority
func NewCertificateAuthority(name string) (*Certificate, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create private key")
	}

	cert, err := NewSelfSignedCACert(name, key)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create self-signed certificate")
	}

	return &Certificate{
		X509:       cert,
		PrivateKey: key,
	}, nil
}

// NewSelfSignedCACert creates a CA certificate.
func NewSelfSignedCACert(name string, key *rsa.PrivateKey) (*x509.Certificate, error) {
	cfg := Config{
		CommonName: name,
	}

	now := time.Now().UTC()

	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		NotBefore:             now,
		NotAfter:              now.Add(duration365d * 10),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		MaxPathLenZero:        true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		IsCA:                  true,
	}

	b, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create self signed CA certificate: %+v", tmpl)
	}

	return x509.ParseCertificate(b)
}

// Certificate is a X509 certifcate / private key pair
type Certificate struct {
	X509       *x509.Certificate
	PrivateKey *rsa.PrivateKey
}

// EncodedCertificate returns PEM-endcoded certificate data.
func (c Certificate) EncodedCertificate() []byte {
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.X509.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// EncodedPrivateKey returns PEM-encoded private key data.
func (c Certificate) EncodedPrivateKey() []byte {
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(c.PrivateKey),
	}
	return pem.EncodeToMemory(&block)
}

// EncodedPublicKey returns PEM-encoded public key data.
func (c Certificate) EncodedPublicKey() []byte {

	publicKey := c.PrivateKey.PublicKey

	der, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}

	if len(der) == 0 {
		panic("nil pub key")
	}

	block := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(&block)
}

// NewCertificate decodes a certificate / private key pair and returns a Certificate
func DecodeCertificate(cert []byte, privateKey []byte) (*Certificate, error) {
	x509, err := decodeCertPEM(cert)
	if err != nil {
		return nil, err
	}
	key, err := decodePrivateKeyPEM(privateKey)
	if err != nil {
		return nil, err
	}
	return &Certificate{
		PrivateKey: key,
		X509:       x509,
	}, nil
}

// GetHash  returns the encoded sha256 hash for the certificate
func (c Certificate) GetHash() (string, error) {
	certHash := sha256.Sum256(c.X509.RawSubjectPublicKeyInfo)
	return "sha256:" + strings.ToLower(hex.EncodeToString(certHash[:])), nil
}

// decodeCertPEM attempts to return a decoded certificate or nil
// if the encoded input does not contain a certificate.
func decodeCertPEM(encoded []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(encoded)
	if block == nil {
		return nil, nil
	}

	return x509.ParseCertificate(block.Bytes)
}

// decodePrivateKeyPEM attempts to return a decoded key or nil
// if the encoded input does not contain a private key.
func decodePrivateKeyPEM(encoded []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(encoded)
	if block == nil {
		return nil, nil
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
