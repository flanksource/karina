# Kind

### Provision kind cluster in docker

```shell
$ platform-cli provision kind-cluster
Creating cluster "kind" ...
 • Ensuring node image (kindest/node:v1.15.7) 🖼  ...
 ✓ Ensuring node image (kindest/node:v1.15.7) 🖼
 • Preparing nodes 📦  ...
 ✓ Preparing nodes 📦
 • Writing configuration 📜  ...
 ✓ Writing configuration 📜
 • Starting control-plane 🕹️  ...
 ✓ Starting control-plane 🕹️
 • Installing StorageClass 💾  ...
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind --kubeconfig /Users/toni/.kube/config

Not sure what to do next? 😅 Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

### Deploy CNI compatible network

```shell
$ platform-cli deploy calico

# Optional if running on Docker for Mac
$ kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true
```

### Deploy the platform

```shell
$ platform-cli deploy stubs
$ platform-cli deploy base
$ platform-cli deploy dex
$ platform-cli deploy minio
...
```