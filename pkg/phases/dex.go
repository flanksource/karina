package phases

import (
	"encoding/base64"
	// "fmt"
	"encoding/json"
	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/runtime/serializer"
	// "crypto/x509"
	"github.com/moshloop/platform-cli/pkg/types"
	// "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/apimachinery/pkg/runtime/serializer/json"
	// "os"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Secret struct {
	meta.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Data contains the secret data. Each key must consist of alphanumeric
	// characters, '-', '_' or '.'. The serialized form of the secret data is a
	// base64 encoded string, representing the arbitrary (possibly non-string)
	// data value here. Described in https://tools.ietf.org/html/rfc4648#section-4
	// +optional
	Data map[string]string `json:"data,omitempty" protobuf:"bytes,2,rep,name=data"`

	// stringData allows specifying non-binary secret data in string form.
	// It is provided as a write-only convenience method.
	// All keys and values are merged into the data field on write, overwriting any existing values.
	// It is never output when reading from the API.
	// +k8s:conversion-gen=false
	// +optional
	StringData map[string]string `json:"stringData,omitempty" protobuf:"bytes,4,rep,name=stringData"`

	// Used to facilitate programmatic handling of secret data.
	// +optional
	Type string `json:"type,omitempty" protobuf:"bytes,3,opt,name=type,casttype=SecretType"`
}

func Dex(platform types.PlatformConfig) error {

	openid := platform.Certificates.OpenID.ToCert()
	log.Infof("Creating dex cert for %s\n", "dex."+platform.Domain)
	cert, err := openid.CreateCertificate("dex."+platform.Domain, "")
	if err != nil {
		return err
	}

	secret := Secret{
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
