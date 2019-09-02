package phases

import (
	"encoding/base64"
	"github.com/moshloop/platform-cli/pkg/api"

	"encoding/json"
	"github.com/moshloop/platform-cli/pkg/types"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

func Dex(platform types.PlatformConfig) error {
	openid := platform.Certificates.OpenID.ToCert()
	log.Infof("Creating dex cert for %s\n", "dex."+platform.Domain)
	cert, err := openid.CreateCertificate("dex."+platform.Domain, "")
	if err != nil {
		return err
	}

	secret := api.Secret{
		TypeMeta:   meta.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: meta.ObjectMeta{Name: "dex-cert", Namespace: "dex"},
		Data: map[string]string{
			"tls.crt": base64.StdEncoding.EncodeToString(cert.EncodedCertificate()),
			"tls.key": base64.StdEncoding.EncodeToString(cert.EncodedPrivateKey()),
		},
		Type: "kubernetes.io/tls",
	}

	data, _ := json.Marshal(secret)
	return ioutil.WriteFile("build/dex-secrets.json", data, 0644)

}
