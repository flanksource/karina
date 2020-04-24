

<h1 align="center">Karina</h1>
  <p align="center">Kubernetes Platform Toolkit</p>
<p align="center">
<a href="https://circleci.com/gh/flanksource/platform-cli"><img src="https://circleci.com/gh/flanksource/platform-cli.svg?style=svg"></a>
<a href="https://goreportcard.com/report/github.com/flanksource/platform-cli"><img src="https://goreportcard.com/badge/github.com/flanksource/platform-cli"></a>
<img src="https://img.shields.io/badge/K8S-1.15%20%7C%201.16-lightgrey.svg"/>
<img src="https://img.shields.io/badge/Infra-vSphere%20%7C%20Kind-lightgrey.svg"/>
<img src="https://img.shields.io/github/license/flanksource/platform-cli.svg?style=flat-square"/>
<a href="https://karina.docs.flanksource.com"> <img src="https://img.shields.io/badge/☰-Docs-lightgrey.svg"/> </a>
</p>

---

**Karina** is a toolkit for building and operating kubernetes based, multi-cluster platforms. It includes the following high level functions:

* **Provisioning** clusters on vSphere and Kind
  * `karina provision`
* **Production Runtime**
  * `karina deploy`
* **Testing Framework** for testing the health of a cluster and the underlying runtime.
  * `karina test`
  * `karina conformance`
* **Rolling Update and Restart** operations
  * `karina rolling restart`
  * `karina rolling update`
* **API/CLI Wrappers** for day-2 operations (backup, restore, configuration) of runtime components including Harbor, Postgres, Consul, Vault and NSX-T/NCP
  * `karina snapshot` dumps specs (excluding secrets), events and logs for troubleshooting
  * `karina logs` exports logs from Elasticsearch using the paging API
  * `karina nsx set-logs` updates runtime logging levels of all nsx components
  * `karina ca generate` create CA key/cert pair suitable for bootstraping
  * `karina kubeconfig` generates kuebconfigs via the master CA or for use with OIDC based login
  * `karina exec` executes a command in every matching pod
  * `karina exec-node` executes a command on every matching node
  * `karina dns` updates DNS
  * `karina db`
  * `karina consul`
  * `karina backup/restore`

#### Cluster Specification

```yaml
# DNS Wildcard domain that this cluster will be accessible under
domain:
# Endpoint for externally hosted consul cluster
consul:
name:
ldap:
  dn:  DC=local,DC=corp
  user: !!env LDAP_USER
  pass: !!env LDAP_PASS
  host:
  adminGroup:
versions:
  kubernetes: v1.17.0
serviceSubnet: 10.96.0.0/16
podSubnet: 10.97.0.0/16
# Prefix to be added to VM hostnames,
hostPrefix:
# The root CA used to sign generated certs
ca:
  cert: .certs/root-ca.crt
  privateKey: .certs/root-ca.key
  password: foobar
 The VM configuration for master nodes
master:
  count: 5
  cpu: 4
  memory: 16
  disk: 200
  network:
  cluster:
  template:
# The VM configuration for worker nodes, multiple groups can be specified
workers:
  worker:
    count: 8
    cpu: 16
    memory: 64
    disk: 300
    network:
 	  cluster:
    template:
```

See [config](https://karina.docs.flanksource.com/reference/config/) for the full list of available fields

#### Production Runtime

* **Docker Registry** (Harbor)
* **Certificate Management** (Cert-Manager)
* **Secret Management** (Sealed Secrets, Vault)
* **Monitoring** (Grafana, Prometheus, Thanos)
* **Logging** (ELK)
* **Authentication** (Dex)
* **Authorization & Policy Enforcement** (OPA)
* **Multi-Tenancy** (Namespace Configurator, Cluster Quotas)
* **Database as a Service** (Postgres)


### Principles

#### Easy for the operator

#### Batteries Included

Functions are integrated but independant, After deploying the production runtime, the testing framework will test and verify, but it can also be used to to components deployed by other mechanisms. Likewise you can provision and deploy, or provision by other means and then deploy the runtime.

#### Escape Hatches

Karina is named after the [Carina Constellation](https://en.wikipedia.org/wiki/Carina_(constellation)) - latin for the hull or keel of a ship.

