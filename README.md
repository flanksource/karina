

<h1 align="center"><img src="https://github.com/flanksource/karina/raw/master/docs/img/logo.png"></i></h1>
  <p align="center">Kubernetes Platform Toolkit</p>
<p align="center">
<img src="https://github.com/flanksource/karina/workflows/E2E/badge.svg"></img>
<img src="https://github.com/flanksource/karina/workflows/Upgrade%20E2E/badge.svg"></img>
<img src="https://github.com/flanksource/karina/workflows/E2E%20vSphere/badge.svg"></img>
<a href="https://goreportcard.com/report/github.com/flanksource/karina"><img src="https://goreportcard.com/badge/github.com/flanksource/karina"></a>
<img src="https://img.shields.io/badge/K8S-1.17%20%7C%201.18-lightgrey.svg"/>
<img src="https://img.shields.io/badge/Infra-vSphere%20%7C%20Kind-lightgrey.svg"/>
<img src="https://img.shields.io/github/license/flanksource/karina.svg?style=flat-square"/>
<a href="https://karina.docs.flanksource.com"> <img src="https://img.shields.io/badge/☰-Docs-lightgrey.svg"/> </a>
<a href="https://join.slack.com/t/flanksource/shared_invite/zt-dvh61tg5-w8XOfrGWtCetGXYk48RKnw"><img src="https://img.shields.io/badge/slack-flanksource-brightgreen.svg?logo=slack"></img></a>
</p>


---

**karina** is a toolkit for building and operating Kubernetes based, multi-cluster platforms. It includes the following high level functions

To see how it compares to other tools in the ecosystem see [comparison](./docs/comparison.md)

<hr>

### Design Principles

* **Batteries Included** - Most components require just a version to enable and are pre-configured with ingress, LDAP and TLS (managed by cert-manager) due to a shared infrastructure model that includes information such as top-level wild card domain, LDAP/S3 connection details, etc.
* **Escape Hatches** for when the defaults don't work for you, easily use kustomize patches to configure resource limits, labels, annotations and anything else on any object managed by karina.
* **Integrated, but independent** - karina works best when used to provision a Kubernetes cluster and then deploy and test a production runtime, but each function can also be used independently, i.e you can run karina e2e tests in an environment that wasn't provisioned or deployed by karina.

### Features

* **Provision** Kubernetes clusters on vSphere (with NSX-T or Calico), Kind and Cluster API (Coming Soon)
* **Deploy** a production runtime for monitoring, logging, security, multi-tenancy, backups, storage, container registry and DBaaS
* **De-Centralized** multi-cluster authentication using a root CA for administrator-level offline authentication, and Dex for online user authentication.
* **CLI Addons/Wrappers** to perform day 2 and incident mitigation tasks such as rolling updates, restarts, backup, restore, failover, replication, logging configuration, system dumps etc.

### Getting Started

To get started provisioning see the quickstart guides for [Kind](https://karina.docs.flanksource.com/admin-guide/provisioning/kind/) and [vSphere](https://karina.docs.flanksource.com/admin-guide/provisioning/vsphere/) <br>

 ### Production Runtime

* **Docker Registry** ([Harbor](http://goharbor.io/))

* **Authentication** ([Dex](https://github.com/dexidp/dex), [Oauth Proxy](https://github.com/oauth2-proxy/oauth2-proxy))
* **Authorization & Policy Enforcement** ([Open Policy Agent ](https://www.openpolicyagent.org/) and [Gatekeeper](https://github.com/open-policy-agent/gatekeeper))

* **Certificate Management** ([cert-manager](https://cert-manager.io/))

- **Secret Management** [(Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets), [Vault](https://www.vaultproject.io/))
- **CI/CD** ([Tekton](https://tekton.dev/), [ArgoCD](https://argoproj.github.io/argo-cd/), [Flux](https://fluxcd.io), [kpack](https://github.com/pivotal/kpack))
- **Database as a Service** ([postgres-operator](https://github.com/zalando/postgres-operator), [rabbitmq-operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html), [redis-operator](https://github.com/spotahome/redis-operator))

* **Logging** (ElasticSearch, Filebeat, Packetbeat, Auditbeat, Kibana)

* **Monitoring** ([Grafana](https://github.com/integr8ly/grafana-operator), [Prometheus](https://github.com/coreos/prometheus-operator), [Thanos](https://thanos.io/), [Karma](https://github.com/prymitive/karma), [Canary Checker](https://github.com/flanksource/canary-checker))

* **Multi-Tenancy** ([Namespace Configurator](https://github.com/redhat-cop/namespace-configuration-operator) Cluster Quotas, [Kiosk](https://github.com/kiosk-sh/kiosk))

### Contributing

Please follow the guideline below when contributing to this project

- [Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)

  