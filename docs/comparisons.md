## Comparisons

#### Production Runtime

A production runtime is the suite of tools that are added to a Kubernetes cluster to provide functionality such as Authentication, Logging and Monitoring

| Runtime                                                      | Description |
| ------------------------------------------------------------ | ----------- |
| [Bitnami Production Runtime](https://github.com/bitnami/kube-prod-runtime) |             |
| [Banzai Pipeline](https://github.com/banzaicloud/pipeline)   |             |
| [Rancher](https://rancher.com/docs/rancher/v2.x/en/overview/) |             |
| [OpenShift](https://www.openshift.com/)                      |             |
| [Lokomotive](https://kinvolk.io/docs/lokomotive/0.6/)        |             |

#### Provisioners

Provisioners are responsible for creating and managing infrastructure and VM's for a Kubernetes cluster to run on.

##### Comparison to other provisioning tools:

| Framework                                                    | Comparison                                                   |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| [Cluster API](https://cluster-api.sigs.k8s.io/)              | Karina is designed to build upon Cluster API, while Cluster API matures much of the control code has been reused, but as soon as Cluster API reaches v1beta1 much of the provisioning code will be replaced with Cluster API. |
| [Rancher](https://rancher.com/docs/rancher/v2.x/en/overview/) |                                                              |
| [Banzai Pipeline](https://github.com/banzaicloud/pipeline)   |                                                              |
| [OpenShift](https://www.openshift.com/)                      |                                                              |
| [kops](https://kops.sigs.k8s.io/)                            |                                                              |
| [kubespray](https://github.com/kubernetes-sigs/kubespray)    |                                                              |
| [lokomotive](https://github.com/kinvolk/lokomotive)          |                                                              |

#### Deployment Tools

##### Comparison to other deployment tools:

| Product                                              | Comparison                                                   |
| ---------------------------------------------------- | ------------------------------------------------------------ |
| [Helm](https://helm.sh/)                             | Helm is the de-facto standard for publishing applications, it is also commonly used for packaging and deploying applications. |
| [Arkade](https://github.com/alexellis/arkade)        |                                                              |
| [reckoner](https://github.com/FairwindsOps/reckoner) |                                                              |
| [helmfile](https://github.com/roboll/helmfile)       |                                                              |
| [helmsman](https://github.com/Praqma/helmsman)       |                                                              |
| [ship](https://github.com/replicatedhq/ship)         |                                                              |
| [tanka](https://tanka.dev/)                          | Flexible, reusable and concise configuration for Kubernetes  |