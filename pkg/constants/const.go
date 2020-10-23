package constants

const KubeSystem = "kube-system"
const PlatformSystem = "platform-system"
const DockerRuntime = "docker"
const ContainerdRuntime = "containerd"

const MasterNodeLabel = "node-role.kubernetes.io/master"
const NodePoolLabel = "karina.flanksource.com/pool"

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
	"vault",
	"velero",
}
