

<h1 align="center"><img src="https://github.com/flanksource/karina/raw/master/docs/img/logo.png"></i></h1>
  <p align="center">Kubernetes Platform Toolkit</p>
<p align="center">
<a href="https://circleci.com/gh/flanksource/karina"><img src="https://circleci.com/gh/flanksource/karina.svg?style=svg"></a>
<a href="https://goreportcard.com/report/github.com/flanksource/karina"><img src="https://goreportcard.com/badge/github.com/flanksource/karina"></a>
<img src="https://img.shields.io/badge/K8S-1.15%20%7C%201.16-lightgrey.svg"/>
<img src="https://img.shields.io/badge/Infra-vSphere%20%7C%20Kind-lightgrey.svg"/>
<img src="https://img.shields.io/github/license/flanksource/karina.svg?style=flat-square"/>
<a href="https://karina.docs.flanksource.com"> <img src="https://img.shields.io/badge/☰-Docs-lightgrey.svg"/> </a>
</p>

---

**karina** is a toolkit for building and operating kubernetes based, multi-cluster platforms. It includes the following high level functions:

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
  * `karina logs` exports logs from ElasticSearch using the paging API
  * `karina nsx set-logs` updates runtime logging levels of all nsx components
  * `karina ca generate` create CA key/cert pair suitable for bootstrapping
  * `karina kubeconfig` generates kuebconfigs via the master CA or for use with OIDC based login
  * `karina exec` executes a command in every matching pod
  * `karina exec-node` executes a command on every matching node
  * `karina dns` updates DNS
  * `karina db`
  * `karina consul`
  * `karina backup/restore`


### Getting Started
To get started provisioning see the quickstart's for [Kind](https://karina.docs.flanksource.com/admin-guide/provisioning/kind.md) and [vSphere](https://karina.docs.flanksource.com/admin-guide/provisioning/vsphere.md) <br>

### Production Runtime

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


