# Cluster Encryption

K8s allows for the configuration of encryption providers for stored secrets (see the [k8s secret storage encryption provider configuration docs](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)).

This is configured by supplying an encryption provider configuration file (look [here](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#encrypting-your-data) for an example).

Karina allows for encryption to be configured at cluster creation as described below.

## Configure Encryption in `config.yaml`.

If an `kubernetes.auditing` section is specified in the config YAML the following configurations can be supplied:

```yaml
kubernetes:
  encryption:
    encryptionProviderConfigFile: ./test/fixtures/encryption-config.yaml
```
The official documentation the [configuration file]((https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#understanding-the-encryption-at-rest-configuration)) and the [various providers](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers) that can be configured.

## Verifying Encryption

The encryption configuration for a specific secret can be verified by using `etcdctl`.

For example in a kind cluster:

```yaml
kubectl exec -n kube-system etcd-kind-control-plane -- \
    /bin/sh -c 'ETCDCTL_API=3 etcdctl get /registry/secrets/default/secret1 \
        --endpoints https://127.0.0.1:2379 \
        --cacert /etc/kubernetes/pki/etcd/ca.crt \
        --cert /etc/kubernetes/pki/etcd/peer.crt \
        --key /etc/kubernetes/pki/etcd/peer.key \
    | strings -n 6 -'
```

an output like 

```
k8s:enc:aescbc:v1:demokey:9
```

shows that the secret `secret1` in namespace `default` has a key `demokey` that is encrypted using the `aescbc` provider.