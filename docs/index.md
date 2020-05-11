![](img/logo.png)

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
To get started provisioning see the quickstart's for [Kind](admin-guide/provisioning/kind.md) and [vSphere](admin-guide/provisioning/vsphere.md) <br>

To see what extensions and add-ons are available to workloads running with the production runtime see the User Guide.

### Principles

#### Easy for the operator

#### Batteries Included

Functions are integrated but independent, After deploying the production runtime, the testing framework will test and verify, but it can also be used to to components deployed by other mechanisms. Likewise you can provision and deploy, or provision by other means and then deploy the runtime.

#### Escape Hatches


### Naming

Karina is named after the [Carina Constellation](https://en.wikipedia.org/wiki/Carina_(constellation)) - latin for the hull or keel of a ship.
