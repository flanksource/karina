package phases

import (
	"encoding/base64"
	"fmt"
	"github.com/moshloop/platform-cli/pkg/platform"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
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
						"idp-certificate-authority-data": string(base64.StdEncoding.EncodeToString([]byte(platform.Certificates.OpenID.X509))),
						"idp-issuer-url":                 "https://dex." + platform.Domain,
					},
				},
			},
		},
		CurrentContext: platform.Name,
	}

	return clientcmd.Write(cfg)
}

func CreateKubeConfig(platform *platform.Platform, endpoint string) ([]byte, error) {
	userName := "kubernetes-admin"
	contextName := fmt.Sprintf("%s@%s", userName, platform.Name)
	cert, err := platform.Certificates.CA.ToCert().CreateCertificate("kubernetes-admin", "system:masters")
	if err != nil {
		return nil, err
	}
	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			platform.Name: {
				Server:                   "https://" + endpoint + ":6443",
				CertificateAuthorityData: []byte(platform.Certificates.CA.X509),
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  platform.Name,
				AuthInfo: userName,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			userName: {
				ClientKeyData:         cert.EncodedPrivateKey(),
				ClientCertificateData: cert.EncodedCertificate(),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.Write(cfg)
}
