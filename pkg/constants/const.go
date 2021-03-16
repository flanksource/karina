package constants

const KubeSystem = "kube-system"
const PlatformSystem = "platform-system"
const DockerRuntime = "docker"
const ContainerdRuntime = "containerd"
const ValidatingWebhookConfiguration = "ValidatingWebhookConfiguration"
const MutatingWebhookConfiguration = "MutatingWebhookConfiguration"
const DefaultIssuer = "default-issuer"
const DefaultIssuerCA = "default-issuer-ca"
const MasterNodeLabel = "node-role.kubernetes.io/master"
const NodePoolLabel = "karina.flanksource.com/pool"
const NodeGroupTaint = "node.kubernetes.io/group"
const ManagedBy = "apps.kubernetes.io/managed-by"
const Karina = "karina"

var PlatformNamespaces = []string{
	"cert-manager",
	"dex",
	"eck",
	"elastic-system",
	"gatekeeper-system",
	"harbor",
	"ingress-nginx",
	"kube-system",
	"local-path-storage",
	"minio",
	"monitoring",
	"nsx-system",
	"opa",
	"platform-system",
	"postgres-operator",
	"quack",
	"sealed-secrets",
	"tekton",
	"vault",
	"velero",
}
