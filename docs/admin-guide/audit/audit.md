# Cluster Auditing

K8s allows for the configuration of auditing on the kube-API-Server (see [k8s audit documentation](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/)). 
This is configured by supplying an audit policy file (look [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml) for an example).
Several other relevant kube-apiserver flags can further configure logging behaviour.

Karina allows for auditing to be configured at cluster creation as described below and supports the use of the log backend.

## Configure Auditing in `config.yaml`.

Karina supports the log backend, which writes audit events to files.

If an `kubernetes.auditing` section is specified in the config YAML the following configurations can be supplied:

```yaml
kubernetes:
  auditing:
    policyFile: ./test/fixtures/audit-policy.yaml
  apiServerExtraArgs:
    "audit-log-path": /var/log/audit/cluster-audit.log
    "audit-log-maxsize": 1024
    "audit-log-maxage": 2
    "audit-log-maxbackup": 3
    "audit-log-format": legacy   # default is json
```
The [official documentation describes](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#log-backend) the `kubeApiServerOptions` parameters.

|Key                   | Description                                              |
|----------------------|----------------------------------------------------------|
| `policyFile`         | Gives the path to the audit policy file to use.<br/>This file will be injected into the master nodes<br/>to `/etc/flanksource/audit-policy/` and into the<br/>api-server pod to `/etc/kubernetes/policies/` and the<br/>api-server `--audit-policy-file` flag will be set to the <br/>correct value.                                           |
| `audit-log-path`     | Specifies the path in the api-server pod that the audit logs are written to. <br>(a value of `'-'` indicates logging to the pod logs.) <br/> If not specified, it defaults to `/var/log/audit/cluster-audit.log`.<br/>Sets the `--audit-log-path` flag.    |
| `audit-log-maxage`   | Specifies the maximum number of days to retain log files.<br/> Sets the `--audit-log-maxage` flag.                      |
| `audit-log-maxbackup`| Specifies the maximum number of audit log files to <br/>retain when logs are rotated past when they reach <br/>maximum size.<br/> Sets the `--audit-log-maxbackup` flag.                   |
| `audit-log-maxsize`  | Specifies the maximum size in megabytes of the audit <br/>log file before it gets rotated <br/>Sets the `--audit-log-maxsize` flag.                   |
| `audit-log-format`   | Specifies the logging format used.<br/>Options are:<br/>`"legacy"` indicates 1-line text format for each event <br/> `"json"` indicates a structured json format. <br/>Sets the `--audit-log-format` flag.                   |

These are injected into the cluster at creation time.

For vSphere clusters these configs are injected through the `kubeadm.conf` passed to `kubeadm` 
through cloud-init images passed to the primary master.

For Kind clusters these configs are injected as `kubeadmConfigPatches` patches specified in the Kind cluster config YAML.

(see [kubeadm control pane configuration documentation](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/control-plane-flags/#apiserver-flags) for more detail)

## Debugging

### Kind Clusters
Investigating the API Server pod spec can indicate its current config:

e.g.

```
kubectl get pod -n kube-system kube-apiserver-kind-control-plane -o yaml
```

Shows the audit spec mapping:

<pre>
       ...

spec:
  containers:
  name: kube-apiserver
  
       ...
       
    volumeMounts:
    - mountPath: <b>/etc/kubernetes/policies/audit-policy.yaml</b>
      name: audit-spec
      readOnly: true
      
       ...

  volumes:
  - hostPath:
      path: <b>/etc/flanksource/audit-policy/audit-policy.yaml</b>
      type: File
    name: audit-spec
    
       ...
</pre>

and the API Server startup flags:

<pre>
     ...
     
spec:
  containers:
  - command:
    - kube-apiserver
    - --advertise-address=172.17.0.2
    - --allow-privileged=true
    - <b>--audit-log-format=json</b>
    - <b>--audit-log-maxage=2</b>
    - <b>--audit-log-maxbackup=3</b>
    - <b>--audit-log-maxsize=10</b>
    - <b>--audit-log-path=/var/log/audit/cluster-audit.log</b>
    - <b>--audit-policy-file=/etc/kubernetes/policies/audit-policy.yaml</b>
    - --authorization-mode=Node,RBAC
    
     ...
</pre>

KIND cluster creation issues can be debugged by specifying the `--trace` argument to `platform-cli` during creation:

e.g.
```bash
platform-cli provision kind-cluster --trace
```
Shows the `kubeadm` patches sent to the KIND configuration and the relevant mappings:

<pre>
<font color="#06989A">INFO</font>[0000] KIND Config YAML:                            
<font color="#06989A">INFO</font>[0000] kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.15.7
  extraMounts:
  
   ...
   
  - containerPath: <b>/etc/flanksource/audit-policy</b>
    hostPath: /code/go/src/github.com/moshloop/platform-cli/test/fixtures
    readOnly: true
 
   ...
 
  kubeadmConfigPatches:
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    kubernetesVersion: v1.15.7
    apiServer:
      timeoutForControlPlane: 4m0s
      extraArgs:
        <b>audit-log-maxage: &quot;2&quot;
        audit-log-maxbackup: &quot;3&quot;
        audit-log-maxsize: &quot;10&quot;
        audit-log-path: /var/log/audit/cluster-audit.log
        audit-policy-file: /etc/kubernetes/policies/audit-policy.yaml</b>
 
    ...
    
      extraVolumes:

    ...

      - name: audit-spec
        hostPath: <b>/etc/flanksource/audit-policy/audit-policy.yaml</b>
        mountPath: <b>/etc/kubernetes/policies/audit-policy.yaml</b>
        readOnly: true
        pathType: File
 
    ...
</pre>

### vSphere Clusters

Adding the following debug keys to the platform config can grant access to the primary master
VM to allow debugging `kubeadm init` config issues:

```YAML
debug:
  injectVspherePrimaryMasterCommand: "useradd debug; echo 'debug:debug' | chpasswd; adduser debug sudo; adduser debug docker;mkdir -p /home/debug/.ssh; touch /home/debug/.ssh/authorized_keys; echo 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDCjSTjgX3GeHQc47Nw1rKF4IwvlR09CncjTsK3GORm9ZpUxPkXhLIQ7xHktYKftapB+zzjfjG02ZtIDwGHYypi5qXLRqPxSLOxjASPIZoErb7WLZ745btEb3pmjBEt19v4fbVFUyr4eqIWzDHGh81Pj2DCuirlMvlWwiHYCiBUsZcRtAlg/u2z4BTfthR2skPvck3Fr3yfL51BHgdv1gdD4n+aAquzxdJV74ED5p9+MKYc7IDkb5NBZf1/8iC3LFw4QjM07ibPc4SDzOMHGRLjCXuEwphfKyv56v1L9lMXXcVrwFSwPCtqQu1uVA2iBufgShq8eWcujLbthfcwP+4v philip@silent' >> /home/debug/.ssh/authorized_keys; chown debug:debug /home/debug/.ssh/authorized_keys; chmod 600 /home/debug/.ssh/authorized_keys"
```

This creates a user named `debug` user (password `debug`) and specifies a public key for a SSH identity that can be used to log into the VM during startup.

cloudinit logs are in `cloud-init-output.log` - the last stages of this contains the kubeadm outputs.

api-server specific logs can be found by finding the api-server pod containers using docker and inspecring their logs.
``````

