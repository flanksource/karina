package kubeadm

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/utils"
	"github.com/moshloop/platform-cli/pkg/api"
	"github.com/moshloop/platform-cli/pkg/platform"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
)

const (
	// AuditPolicyPath is the fixed location where kubernetes cluster audit policy files are placed.
	AuditPolicyPath              = "/etc/kubernetes/policies/audit-policy.yaml"
	EncryptionProviderConfigPath = "/etc/kubernetes/policies/encryption-provider-config.yaml" //TODO: verify location
)

// NewClusterConfig constructs a default new ClusterConfiguration from a given Platform config
func NewClusterConfig(cfg *platform.Platform) api.ClusterConfiguration {
	cluster := api.ClusterConfiguration{
		APIVersion:        "kubeadm.k8s.io/v1beta2",
		Kind:              "ClusterConfiguration",
		KubernetesVersion: cfg.Kubernetes.Version,
		CertificatesDir:   "/etc/kubernetes/pki",
		ClusterName:       cfg.Name,
		ImageRepository:   "k8s.gcr.io",
		// Control plane endpoint is load balanced client side using haproxy + consul service discovery
		ControlPlaneEndpoint: "localhost:8443",
	}
	cluster.Networking.DNSDomain = "cluster.local"
	cluster.Networking.ServiceSubnet = cfg.ServiceSubnet
	cluster.Networking.PodSubnet = cfg.PodSubnet
	cluster.DNS.Type = "CoreDNS"
	cluster.Etcd.Local.DataDir = "/var/lib/etcd"
	cluster.Etcd.Local.ExtraArgs = cfg.Kubernetes.EtcdExtraArgs
	cluster.Etcd.Local.ExtraArgs["listen-metrics-urls"] = "http://0.0.0.0:2381"

	cluster.APIServer.CertSANs = []string{"localhost", "127.0.0.1", "k8s-api." + cfg.Domain}
	cluster.APIServer.TimeoutForControlPlane = "4m0s"
	cluster.APIServer.ExtraArgs = cfg.Kubernetes.APIServerExtraArgs

	if cfg.Kubernetes.AuditConfig.PolicyFile != "" {
		cluster.APIServer.ExtraArgs["audit-policy-file"] = AuditPolicyPath
		mnt := api.HostPathMount{
			Name:      "auditpolicy",
			HostPath:  AuditPolicyPath,
			MountPath: AuditPolicyPath,
			ReadOnly:  true,
			PathType:  api.HostPathFile,
		}
		cluster.APIServer.ExtraVolumes = append(cluster.APIServer.ExtraVolumes, mnt)
	}

	if cfg.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile != "" {
		cluster.APIServer.ExtraArgs["encryption-provider-config"] = EncryptionProviderConfigPath
		mnt := api.HostPathMount{
			Name:      "encryption-config",
			HostPath:  EncryptionProviderConfigPath,
			MountPath: EncryptionProviderConfigPath,
			ReadOnly:  true,
			PathType:  api.HostPathFile,
		}
		cluster.APIServer.ExtraVolumes = append(cluster.APIServer.ExtraVolumes, mnt)
	}

	if !cfg.Ldap.Disabled {
		cluster.APIServer.ExtraArgs["oidc-issuer-url"] = "https://dex." + cfg.Domain
		cluster.APIServer.ExtraArgs["oidc-client-id"] = "kubernetes"
		cluster.APIServer.ExtraArgs["oidc-ca-file"] = "/etc/ssl/certs/openid-ca.pem"
		cluster.APIServer.ExtraArgs["oidc-username-claim"] = "email"
		cluster.APIServer.ExtraArgs["oidc-groups-claim"] = "groups"
	}
	if strings.HasPrefix(cluster.KubernetesVersion, "v1.16") {
		runtimeConfigs := []string{
			"apps/v1beta1=true",
			"apps/v1beta2=true",
			"extensions/v1beta1/daemonsets=true",
			"extensions/v1beta1/deployments=true",
			"extensions/v1beta1/replicasets=true",
			"extensions/v1beta1/networkpolicies=true",
			"extensions/v1beta1/podsecuritypolicies=true",
		}
		cluster.APIServer.ExtraArgs["runtime-config"] = strings.Join(runtimeConfigs, ",")
	}
	cluster.ControllerManager.ExtraArgs = cfg.Kubernetes.ControllerExtraArgs
	cluster.ControllerManager.ExtraArgs["cluster-signing-cert-file"] = "/etc/kubernetes/pki/csr-ca.crt"
	cluster.ControllerManager.ExtraArgs["cluster-signing-key-file"] = "/etc/kubernetes/pki/ca.key"
	cluster.Scheduler.ExtraArgs = cfg.Kubernetes.SchedulerExtraArgs
	return cluster
}

func NewInitConfig(cfg *platform.Platform) api.InitConfiguration {
	config := api.InitConfiguration{
		Kind: "InitConfiguration",
		NodeRegistration: api.NodeRegistration{
			KubeletExtraArgs: cfg.Kubernetes.KubeletExtraArgs,
		},
	}

	return config
}

// createBootstrapToken is extracted from https://github.com/kubernetes-sigs/cluster-api-bootstrap-provider-kubeadm/blob/master/controllers/token.go
func CreateBootstrapToken(client corev1.SecretInterface) (string, error) {
	// createToken attempts to create a token with the given ID.
	token, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return "", fmt.Errorf("unable to generate bootstrap token: %v", err)
	}

	substrs := bootstraputil.BootstrapTokenRegexp.FindStringSubmatch(token)
	if len(substrs) != 3 {
		return "", fmt.Errorf("the bootstrap token %q was not of the form %q", token, bootstrapapi.BootstrapTokenPattern)
	}
	tokenID := substrs[1]
	tokenSecret := substrs[2]

	secretName := bootstraputil.BootstrapTokenSecretName(tokenID)
	secretToken := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: bootstrapapi.SecretTypeBootstrapToken,
		Data: map[string][]byte{
			bootstrapapi.BootstrapTokenIDKey:               []byte(tokenID),
			bootstrapapi.BootstrapTokenSecretKey:           []byte(tokenSecret),
			bootstrapapi.BootstrapTokenExpirationKey:       []byte(time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)),
			bootstrapapi.BootstrapTokenUsageSigningKey:     []byte("true"),
			bootstrapapi.BootstrapTokenUsageAuthentication: []byte("true"),
			bootstrapapi.BootstrapTokenExtraGroupsKey:      []byte("system:bootstrappers:kubeadm:default-node-token"),
			bootstrapapi.BootstrapTokenDescriptionKey:      []byte("token generated by platform-cli"),
		},
	}

	if _, err = client.Create(secretToken); err != nil {
		return "", err
	}
	return token, nil
}

func UploadControlPaneCerts(platform *platform.Platform) (string, error) {
	client, err := platform.GetClientset()
	if err != nil {
		return "", err
	}

	masterNode, err := platform.GetMasterNode()

	if err != nil {
		return "", err
	}

	secrets := client.CoreV1().Secrets("kube-system")
	var key string
	secret, err := secrets.Get("kubeadm-certs", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		key = utils.RandomKey(32)
		platform.Infof("Uploading control plane cert from %s", masterNode)
		stdout, err := platform.Executef(masterNode, 2*time.Minute, "kubeadm init phase upload-certs --upload-certs --skip-certificate-key-print --certificate-key %s", key)
		platform.Infof("Uploaded control plane certs: %s (%v)", stdout, err)
		secret, err = secrets.Get("kubeadm-certs", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		// FIXME storing the encryption key in plain text alongside the certs, kind of defeats the purpose
		secret.Annotations = map[string]string{"key": key}
		if _, err := secrets.Update(secret); err != nil {
			return "", err
		}
		return key, nil
	} else if err == nil {
		platform.Infof("Found existing control plane certs created: %v", secret.GetCreationTimestamp())
		return secret.Annotations["key"], nil
	}
	return "", err
}

func GetOrCreateBootstrapToken(platform *platform.Platform) (string, error) {
	if platform.BootstrapToken != "" {
		return platform.BootstrapToken, nil
	}
	client, err := platform.GetClientset()
	if err != nil {
		return "", err
	}
	token, err := CreateBootstrapToken(client.CoreV1().Secrets("kube-system"))
	if err != nil {
		return "", err
	}
	platform.BootstrapToken = token

	return platform.BootstrapToken, nil
}
