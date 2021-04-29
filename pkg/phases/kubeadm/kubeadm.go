package kubeadm

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/api"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"gopkg.in/flanksource/yaml.v3"
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
	EncryptionProviderConfigPath = "/etc/kubernetes/policies/encryption-provider-config.yaml"
	CSRCAPath                    = "/etc/kubernetes/pki/csr-ca.crt"
	CSRKeyPath                   = "/etc/kubernetes/pki/csr-ca.key"
	noCAErrorText                = "must specify a ca"
)

// NewClusterConfig constructs a default new ClusterConfiguration from a given Platform config
func NewClusterConfig(cfg *platform.Platform) api.ClusterConfiguration {
	cluster := api.ClusterConfiguration{
		APIVersion:           "kubeadm.k8s.io/v1beta2",
		Kind:                 "ClusterConfiguration",
		KubernetesVersion:    cfg.Kubernetes.Version,
		CertificatesDir:      "/etc/kubernetes/pki",
		ClusterName:          cfg.Name,
		ImageRepository:      "k8s.gcr.io",
		ControlPlaneEndpoint: cfg.JoinEndpoint,
	}
	cluster.Networking.DNSDomain = "cluster.local"
	cluster.Networking.ServiceSubnet = cfg.ServiceSubnet
	cluster.Networking.PodSubnet = cfg.PodSubnet
	cluster.DNS.Type = "CoreDNS"
	cluster.Etcd.Local.DataDir = "/var/lib/etcd"
	cluster.Etcd.Local.ExtraArgs = cfg.Kubernetes.EtcdExtraArgs
	cluster.Etcd.Local.ExtraArgs["listen-metrics-urls"] = "http://0.0.0.0:2381"
	cluster.APIServer.CertSANs = []string{"localhost", "127.0.0.1", "k8s-api." + cfg.Domain}
	cluster.APIServer.TimeoutForControlPlane = "10m0s"
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

	if !cfg.Ldap.Disabled && cfg.IngressCA != nil {
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
	cluster.ControllerManager.ExtraArgs["bind-address"] = "0.0.0.0"
	cluster.ControllerManager.ExtraArgs["cluster-signing-cert-file"] = CSRCAPath
	cluster.ControllerManager.ExtraArgs["cluster-signing-key-file"] = CSRKeyPath
	cluster.Scheduler.ExtraArgs = cfg.Kubernetes.SchedulerExtraArgs
	cluster.Scheduler.ExtraArgs["bind-address"] = "0.0.0.0"
	return cluster
}

func GetFilesToMountForPrimary(platform *platform.Platform) (map[string]string, error) {
	// the primary or first control plane node gets everything subsequent nodes do

	files, err := GetFilesToMountForSecondary(platform)
	if err != nil {
		return nil, err
	}

	// plus the initial certificates for bootstrapping, subsequent control plane nodes download them from a secret

	if platform.CA == nil {
		return nil, fmt.Errorf(noCAErrorText)
	}
	clusterCA := certs.NewCertificateBuilder("kubernetes-ca").CA().Certificate
	clusterCA, err = platform.GetCA().SignCertificate(clusterCA, 10)
	if err != nil {
		return nil, fmt.Errorf("addCerts: failed to sign certificate: %v", err)
	}

	// Because the certificate-signing-cert can only contain a single certificate and  ca.crt contains 2 (the root/global ca for admin auth, and the cluster ca for node auth)
	// ca.{key,crt} are copied removing the 2nd cert and specified as extra arguments to the api server, because of this kubeadm upload/download certs
	// is not aware of these files and they get recreated for each new master, causing node join issues down the line, we therefore split the certs
	crt := string(clusterCA.EncodedCertificate()) + "\n"
	crt = crt + string(platform.GetCA().GetPublicChain()[0].EncodedCertificate()) + "\n"
	files[CSRCAPath] = string(clusterCA.EncodedCertificate())
	files[CSRKeyPath] = string(clusterCA.EncodedPrivateKey())
	files["/etc/kubernetes/pki/ca.crt"] = crt
	files["/etc/kubernetes/pki/ca.key"] = string(clusterCA.EncodedPrivateKey())
	return files, nil
}

func GetFilesToMountForSecondary(platform *platform.Platform) (map[string]string, error) {
	var files = make(map[string]string)
	files["/etc/ssl/certs/openid-ca.pem"] = string(platform.GetIngressCA().GetPublicChain()[0].EncodedCertificate())

	if platform.Kubernetes.AuditConfig.PolicyFile != "" {
		contents, err := ioutil.ReadFile(platform.Kubernetes.AuditConfig.PolicyFile)
		if err != nil {
			return nil, err
		}
		files[AuditPolicyPath] = string(contents)
	}

	if platform.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile != "" {
		contents, err := ioutil.ReadFile(platform.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile)
		if err != nil {
			return nil, err
		}
		files[EncryptionProviderConfigPath] = string(contents)
	}

	return files, nil
}

func getKubeletArgs(cfg *platform.Platform) map[string]string {
	args := cfg.Kubernetes.KubeletExtraArgs
	if cfg.Vsphere != nil && !cfg.Vsphere.IsDisabled() && cfg.Vsphere.CPIVersion != "" {
		if args == nil {
			args = make(map[string]string)
		}
		args["cloud-provider"] = "external"
	}
	return args
}

func NewInitConfig(cfg *platform.Platform) api.InitConfiguration {
	return api.InitConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta2",
		Kind:       "InitConfiguration",
		NodeRegistration: api.NodeRegistration{
			KubeletExtraArgs: getKubeletArgs(cfg),
		},
	}
}

func NewControlPlaneJoinConfiguration(cfg *platform.Platform) ([]byte, error) {
	token, err := GetOrCreateBootstrapToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create bootstrap token: %v", err)
	}
	certKey, err := UploadControlPlaneCerts(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to upload control plane certs: %v", err)
	}
	configuration := api.JoinConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta2",
		Kind:       "JoinConfiguration",
		ControlPlane: &api.JoinControlPlane{
			CertificateKey: certKey,
			LocalAPIEndpoint: api.APIEndpoint{
				AdvertiseAddress: "0.0.0.0",
				BindPort:         6443,
			},
		},
		Discovery: api.Discovery{
			BootstrapToken: &api.BootstrapTokenDiscovery{
				APIServerEndpoint:        cfg.JoinEndpoint,
				Token:                    token,
				UnsafeSkipCAVerification: true,
			},
		},
		NodeRegistration: api.NodeRegistration{
			KubeletExtraArgs: getKubeletArgs(cfg),
		},
	}
	if cfg.Kubernetes.ContainerRuntime == constants.ContainerdRuntime {
		configuration.NodeRegistration.CRISocket = "unix:///run/containerd/containerd.sock"
	}
	return yaml.Marshal(configuration)
}

func NewJoinConfiguration(cfg *platform.Platform, node types.VM) ([]byte, error) {
	token, err := GetOrCreateBootstrapToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create bootstrap token: %v", err)
	}

	kubeletExtraArgs := getKubeletArgs(cfg)
	if node.KubeletExtraArgs != nil {
		for k, v := range node.KubeletExtraArgs {
			kubeletExtraArgs[k] = v
		}
	}

	configuration := api.JoinConfiguration{
		APIVersion: "kubeadm.k8s.io/v1beta2",
		Kind:       "JoinConfiguration",
		NodeRegistration: api.NodeRegistration{
			KubeletExtraArgs: kubeletExtraArgs,
		},
		Discovery: api.Discovery{
			BootstrapToken: &api.BootstrapTokenDiscovery{
				APIServerEndpoint:        cfg.JoinEndpoint,
				Token:                    token,
				UnsafeSkipCAVerification: true,
			},
		},
	}
	if cfg.Kubernetes.ContainerRuntime == constants.ContainerdRuntime {
		configuration.NodeRegistration.CRISocket = "unix:///run/containerd/containerd.sock"
	}
	return yaml.Marshal(configuration)
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
			bootstrapapi.BootstrapTokenExpirationKey:       []byte(time.Now().Add(24 * time.Hour).Format(time.RFC3339)),
			bootstrapapi.BootstrapTokenUsageSigningKey:     []byte("true"),
			bootstrapapi.BootstrapTokenUsageAuthentication: []byte("true"),
			bootstrapapi.BootstrapTokenExtraGroupsKey:      []byte("system:bootstrappers:kubeadm:default-node-token"),
			bootstrapapi.BootstrapTokenDescriptionKey:      []byte("token generated by karina"),
		},
	}

	if _, err = client.Create(context.TODO(), secretToken, metav1.CreateOptions{}); err != nil {
		return "", err
	}
	return token, nil
}

func UploadEtcdCerts(platform *platform.Platform) (*certs.Certificate, error) {
	client, err := platform.GetClientset()
	if err != nil {
		return nil, err
	}

	masterNode, err := platform.GetMasterNode()
	if err != nil {
		return nil, err
	}

	secrets := client.CoreV1().Secrets("kube-system")
	secret, err := secrets.Get(context.TODO(), "etcd-certs", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		platform.Infof("Uploading etcd certs from %s", masterNode)
		stdout, err := platform.Executef(masterNode, 2*time.Minute, "kubectl --kubeconfig /etc/kubernetes/admin.conf -n kube-system create secret tls etcd-certs --cert=/etc/kubernetes/pki/etcd/ca.crt --key=/etc/kubernetes/pki/etcd/ca.key")
		platform.Infof("Uploaded control plane certs: %s (%v)", stdout, err)
		secret, err = secrets.Get(context.TODO(), "etcd-certs", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return certs.DecodeCertificate(secret.Data["tls.crt"], secret.Data["tls.key"])
}

func UploadControlPlaneCerts(platform *platform.Platform) (string, error) {
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
	secret, err := secrets.Get(context.TODO(), "kubeadm-certs", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		key = utils.RandomKey(32)
		platform.Infof("Uploading control plane cert from %s", masterNode)
		stdout, err := platform.Executef(masterNode, 2*time.Minute, "kubeadm init phase upload-certs --upload-certs --skip-certificate-key-print --certificate-key %s", key)
		platform.Infof("Uploaded control plane certs: %s (%v)", stdout, err)
		secret, err = secrets.Get(context.TODO(), "kubeadm-certs", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		// FIXME storing the encryption key in plain text alongside the certs, kind of defeats the purpose
		secret.Annotations = map[string]string{"key": key}
		if _, err := secrets.Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			return "", err
		}
		return key, nil
	} else if err == nil {
		platform.Debugf("Found existing control plane certs created: %v", secret.GetCreationTimestamp())
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

func GetClusterVersion(platform *platform.Platform) (string, error) {
	config := api.ClusterConfiguration{}
	data := (*platform.GetConfigMap("kube-system", "kubeadm-config"))["ClusterConfiguration"]
	if err := yaml.Unmarshal([]byte(data), &config); err != nil {
		return "", err
	}
	return config.KubernetesVersion, nil
}

func GetNodeVersion(platform *platform.Platform, node v1.Node) string {
	client, err := platform.GetClientset()
	if err != nil {
		return "<err>"
	}
	pods, err := client.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
		LabelSelector: "component=kube-apiserver",
	})
	if err != nil {
		return "<err>"
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			return strings.Split(container.Image, ":")[1]
		}
	}
	return "?"
}
