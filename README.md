

<h1 align="center"><img src="https://github.com/flanksource/karina/raw/master/docs/img/logo.png"></i></h1>
  <p align="center">Kubernetes Platform Toolkit</p>
<p align="center">
<a href="https://circleci.com/gh/flanksource/karina"><img src="https://circleci.com/gh/flanksource/karina.svg?style=svg"></a>
<a href="https://goreportcard.com/report/github.com/flanksource/karina"><img src="https://goreportcard.com/badge/github.com/flanksource/karina"></a>
<img src="https://img.shields.io/badge/K8S-1.17%20%7C%201.18-lightgrey.svg"/>
<img src="https://img.shields.io/badge/Infra-vSphere%20%7C%20Kind-lightgrey.svg"/>
<img src="https://img.shields.io/github/license/flanksource/karina.svg?style=flat-square"/>
<a href="https://karina.docs.flanksource.com"> <img src="https://img.shields.io/badge/☰-Docs-lightgrey.svg"/> </a>
</p>


---

**karina** is a toolkit for building and operating Kubernetes based, multi-cluster platforms. It includes the following high level functions:



### Features

* **Provision** Kubernetes clusters on vSphere (with NSX-T or Calico), Kind and Cluster API (Coming Soon)
* **Deploy** a production runtime for monitoring, logging, security, multi-tenancy, backups, storage, container registry and DBaaS
* **Batteries Included** - Most components require just a version to enable and are pre-configured with ingress, LDAP and TLS (managed by cert-manager) due to a shared infrastructure model that includes information such as top-level wild card domain, LDAP connection details, S3 connection details, etc.
* **Escape Hatches** for when the defaults don't work for you, easily use kustomize patches to configure resource limits, labels, annotations and arguments.
* **Integrated, but independent** - karina works best when used to provision a Kubernetes cluster and then deploy and test a production runtime, but each function can also be used independently, i.e you can run karina e2e tests in an environment that wasn't provisioned or deployed by karina.
* **De-Centralized** multi-cluster authentication using a root CA for administrator-level offline authentication, and Dex for online user authentication.
CLI Clients for
* **CLI Addons/Wrappers** to perform day 2 and incident mitigation tasks such as rolling updates, restarts, backup, restore, failover, replication, logging configuration, system dumps etc.



### Getting Started
To get started provisioning see the quickstart guides for [Kind](https://karina.docs.flanksource.com/admin-guide/provisioning/kind/) and [vSphere](https://karina.docs.flanksource.com/admin-guide/provisioning/vsphere/) <br>

### Production Runtime

* **Docker Registry** ([Harbor](http://goharbor.io/))
* **Certificate Management** ([cert-manager](https://cert-manager.io/))
* **Secret Management** [(Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets), [Vault](https://www.vaultproject.io/))
* **Monitoring** ([Grafana](https://github.com/integr8ly/grafana-operator), [Prometheus](https://github.com/coreos/prometheus-operator), [Thanos](https://thanos.io/), [Karma](https://github.com/prymitive/karma), [Canary Checker](https://github.com/flanksource/canary-checker))
* **Logging** (ElasticSearch, Filebeat, Packetbeat, Auditbeat, Kibana)
* **Authentication** ([Dex](https://github.com/dexidp/dex))
* **Authorization & Policy Enforcement** ([Open Policy Agent](https://www.openpolicyagent.org/))
* **Multi-Tenancy** ([Namespace Configurator](https://github.com/redhat-cop/namespace-configuration-operator) Cluster Quotas, [Kiosk](https://github.com/kiosk-sh/kiosk))
* **Database as a Service** ([postgres-operato](https://github.com/zalando/postgres-operator)r)



## Comparisons

#### Production Runtime

A production runtime is the suite of tools that are added to a Kubernetes cluster to provide functionality such as Authentication, Logging and Monitoring

| Runtime                                                                     | Description |
| --------------------------------------------------------------------------- | ----------- |
| [Bitnami Production Runtime](https://github.com/bitnami/kube-prod-runtime)  |             |
| [Banzai Pipeline](https://github.com/banzaicloud/pipeline)                  |             |
| [Rancher](https://rancher.com/docs/rancher/v2.x/en/overview/)               |             |
| [OpenShift](https://www.openshift.com/)                                     |             |
| [Lokomotive](https://kinvolk.io/docs/lokomotive/0.6/)                       |             | 

#### Provisioners

Provisioners are responsible for creating and managing infrastructure and VM's for a Kubernetes cluster to run on.

| Framework                                                    | Comparison                                                   |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| [Cluster API](https://cluster-api.sigs.k8s.io/)              | Karina is designed to build upon Cluster API, while Cluster API matures much of the control code has been reused, but as soon as Cluster API reaches v1beta1 much of the provisioning code will be replaced with Cluster API. |
| [Rancher](https://rancher.com/docs/rancher/v2.x/en/overview/) |                                                              |
| [Banzai Pipeline](https://github.com/banzaicloud/pipeline)   |                                                              |
| [OpenShift](https://www.openshift.com/)                      |                                                              |
| [kops](https://kops.sigs.k8s.io/)                            |                                                              |
| [kubespray](https://github.com/kubernetes-sigs/kubespray)    |                                                              |
| [Lokomotive](https://kinvolk.io/docs/lokomotive/0.6/concepts/components/)

#### Deployment Tools

| Product     | Comparison                                                   |
| ----------- | ------------------------------------------------------------ |
| [Helm](https://helm.sh/) | Helm is the de-facto standard for publishing applications, it is also commonly used for packaging and deploying applications. |
| [Arkade](https://github.com/alexellis/arkade) |                                                              |
|[reckoner](https://github.com/FairwindsOps/reckoner)       |                                                              |
|[helmfile](https://github.com/roboll/helmfile)  | |
| [helmsman](https://github.com/Praqma/helmsman) | |
| [ship](https://github.com/replicatedhq/ship) | |



##### Where does the name come from?

Karina is named after the [Carina Constellation](https://en.wikipedia.org/wiki/Carina_(constellation)) - Latin for the hull or keel of a ship.

