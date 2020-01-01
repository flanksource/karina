package phases

import (
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/moshloop/commons/certs"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func CreateOIDCKubeConfig(platform *platform.Platform, endpoint string) ([]byte, error) {
	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			platform.Name: {
				Server:                "https://" + endpoint + ":6443",
				InsecureSkipTLSVerify: true,
			},
		},
		Contexts: map[string]*api.Context{
			platform.Name: {
				Cluster:  platform.Name,
				AuthInfo: "sso",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"sso": {
				AuthProvider: &api.AuthProviderConfig{
					Name: "oidc",
					Config: map[string]string{
						"client-id":                      "kubernetes",
						"client-secret":                  "ZXhhbXBsZS1hcHAtc2VjcmV0",
						"extra-scopes":                   "offline_access openid profile email groups",
						"idp-certificate-authority-data": string(base64.StdEncoding.EncodeToString([]byte(platform.GetIngressCA().GetPublicChain()[0].EncodedCertificate()))),
						"idp-issuer-url":                 "https://dex." + platform.Domain,
					},
				},
			},
		},
		CurrentContext: platform.Name,
	}

	return clientcmd.Write(cfg)
}

func CreateKubeConfig(platform *platform.Platform, endpoint, group, user string) ([]byte, error) {
	contextName := fmt.Sprintf("%s@%s", user, platform.Name)
	cert := certs.NewCertificateBuilder(group).Client().Certificate
	cert, err := platform.GetCA().SignCertificate(cert, 1)
	if err != nil {
		return nil, err
	}
	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			platform.Name: {
				Server:                "https://" + endpoint + ":6443",
				InsecureSkipTLSVerify: true,
				// CertificateAuthorityData: []byte(platform.Certificates.CA.X509),
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  platform.Name,
				AuthInfo: contextName,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			contextName: {
				ClientKeyData:         cert.EncodedPrivateKey(),
				ClientCertificateData: cert.EncodedCertificate(),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.Write(cfg)
}
