package phases

import (
	// "crypto/x509"
	"github.com/moshloop/platform-cli/pkg/types"
	// "github.com/moshloop/platform-cli/pkg/utils"
	// "os"
	// // log "github.com/sirupsen/logrus"
	// "io/ioutil"
)

func Dex(cfg types.PlatformConfig) error {

	// cert := utils.Config{
	// 	CommonName: "dex." + cfg.Domain,
	// 	Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	// }
	// key, _ := utils.NewPrivateKey()
	// ca, _ := utils.NewSelfSignedCACert(key)
	// server, err := cert.NewSignedCert(key, ca, key)
	// if err != nil {
	// 	return err
	// }
	// os.MkdirAll("build/secrets", 0755)

	// if err := ioutil.WriteFile("build/secrets/dex.key", utils.EncodePrivateKeyPEM(key), 0644); err != nil {
	// 	return err
	// }
	// if err := ioutil.WriteFile("build/secrets/ca.pem", utils.EncodeCertPEM(ca), 0644); err != nil {
	// 	return err
	// }
	// if err := ioutil.WriteFile("build/secrets/server.pem", utils.EncodeCertPEM(server), 0644); err != nil {
	// 	return err
	// }
	return nil
}
