
![](../../img/Cluster%20Lifecycle.png)

##### Control Plane Init

The control plane is initialized by:

- Generating all the certificates (to be used to generate access credentials or provision new control plane nodes)
- Injecting the certificates and configs into a cloud-config file and a provisioning a VM, on boot `kubeadm init` is run to bootstrap the cluster

##### Adding Secondary Control Plane Nodes

- The certificates are injected into cloud-init and multiple VM's are provisioned concurrently which run `kubeadm join --control-plane` on boot

##### Adding Workers

- Workers have a bootstrap token injected into cloud-init and multiple VM's are provisioned concurrently which run `kubeadm --join` on boot
