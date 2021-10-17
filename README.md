

<h1 align="center"><img src="https://github.com/flanksource/karina/raw/master/docs/img/logo.png"></i></h1>
  <p align="center">Kubernetes Platform Toolkit</p>
<p align="center">
<a href="https://goreportcard.com/report/github.com/flanksource/karina"><img src="https://goreportcard.com/badge/github.com/flanksource/karina"></a>
<img src="https://img.shields.io/badge/Infra-vSphere%20%7C%20Kind-lightgrey.svg"/>
<img src="https://img.shields.io/github/license/flanksource/karina.svg?style=flat-square"/>
<a href="https://karina.docs.flanksource.com"> <img src="https://img.shields.io/badge/☰-Docs-lightgrey.svg"/> </a>
<a href="https://join.slack.com/t/flanksource/shared_invite/zt-dvh61tg5-w8XOfrGWtCetGXYk48RKnw"><img src="https://img.shields.io/badge/slack-flanksource-brightgreen.svg?logo=slack"></img></a>
</p>

---

**karina**  is an open-source toolkit for building platforms, which includes:

* **Integrated Production Runtime** for monitoring, logging, multi-tenancy, auth, policy, secrets, DBaaS and CICD.
  * **Docker Registry** ([Harbor](http://goharbor.io/))
  * **Authentication** ([Dex](https://github.com/dexidp/dex), [Oauth Proxy](https://github.com/oauth2-proxy/oauth2-proxy))
  * **Authorization & Policy Enforcement** ([Open Policy Agent ](https://www.openpolicyagent.org/) and [Gatekeeper](https://github.com/open-policy-agent/gatekeeper))
  * **Certificate Management** ([cert-manager](https://cert-manager.io/))
  * **Secret Management** [(Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets), [Vault](https://www.vaultproject.io/))
  * **CI/CD** ([Tekton](https://tekton.dev/), [ArgoCD](https://argoproj.github.io/argo-cd/), [Flux](https://fluxcd.io), [kpack](https://github.com/pivotal/kpack), [keptn](https://github.com/keptn/keptn))
  * **Database as a Service** ([postgres-operator](https://github.com/zalando/postgres-operator), [rabbitmq-operator](https://www.rabbitmq.com/kubernetes/operator/operator-overview.html), [redis-operator](https://github.com/spotahome/redis-operator))
  * **Logging** (ElasticSearch, Filebeat, Packetbeat, Auditbeat, Kibana)
  * **Monitoring** ([Grafana](https://github.com/integr8ly/grafana-operator), [Prometheus](https://github.com/coreos/prometheus-operator), [Thanos](https://thanos.io/), [Karma](https://github.com/prymitive/karma), [Canary Checker](https://github.com/flanksource/canary-checker))
  * **Multi-Tenancy** ([Namespace Configurator](https://github.com/redhat-cop/namespace-configuration-operator) Cluster Quotas, [Kiosk](https://github.com/kiosk-sh/kiosk))
* **Cluster Provisioning** framework for [Kind](https://karina.docs.flanksource.com/admin-guide/provisioning/kind/), [vSphere](https://karina.docs.flanksource.com/admin-guide/provisioning/vsphere/) and Cluster API (Coming Soon)
* **Operations focused CLI** for health checks, etcd maintenance, backup and restore, rolling updates, logging, etc..

karina leverages a number of other standalone operators built by flanksource:

* [kommons](https://github.com/flanksource/kommons) as high-level library wrapping client-go
* [canary-checker](https://github.com/flanksource/canary-checker) for multi-cluster synthetic monitoring of pod scheduling, docker pulls, DNS, ICMP, LDAP, Postgres, HTTP, etc..
* [template-operator](https://github.com/flanksource/template-operator) for building re-usable CRD's similar to crossplane's [XRD](https://crossplane.io/docs/v0.14/introduction/composition.html)
* [platform-operator](https://github.com/flanksource/platform-operator) for multi-tenancy support
* [git-operator](https://github.com/flanksource/git-operator) for deploying GitOps tools like Flux and exposing REST interfaces onto of Git.

<hr>

### Design Principles

* **Batteries Included** - Most components require just a version to enable and are pre-configured with ingress, LDAP and TLS (managed by cert-manager) due to a shared infrastructure model that includes information such as top-level wild card domain, LDAP/S3 connection details, etc.
* **Escape Hatches** for when the defaults don't work for you, easily use kustomize patches to configure resource limits, labels, annotations and anything else on any object managed by karina.
* **Integrated, but independent** - karina works best when used to provision a Kubernetes cluster and then deploy and test a production runtime, but each function can also be used independently, i.e you can run karina e2e tests in an environment that wasn't provisioned or deployed by karina.

### Comparisons

To see how karina compares to other tools in the ecosystem see [comparisons](./docs/comparisons.md)

### Contributing

Please follow the guideline below when contributing to this project

- [Conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
