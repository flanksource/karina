
## Incident Roles

|                                |                                 |
| ------------------------------ | ------------------------------- |
| :worker: Incident Responder(s) | Hands on keyboard               |
| :captain: Incident Commander   | Communicating with stakeholders |
| :talk: Incident Communicator   | Overseeing the incident         |

The steps below are applicable for Severity ("**Sev**") 1 & 2 level incidents. See the [Escalations](#escalations) section for more detailed steps with times. See [Severity Classification](#severity-classification) for a severity classification rubric.

:1: The first responder automatically becomes the incident commander until the role is transferred:

!!! warning
    You are still the :captain: commander until someone else says  :speak:  *I have command*

:2: The incident commander is automatically the incident communicator until it is transferred:

!!! warning
    You are still :talk: communicator until someone else says :speak:  *I have comms*

:3: Set up a communications channel on %%{support.channel}%%

:4: Communicate a new incident and channel details to %%{support.email}%%

## Severity Classification

<table>
  <tr>
   <td>
   </td>
   <td>Sev 1
   </td>
   <td>Sev 2
   </td>
   <td>Sev 3
   </td>
  </tr>
  <tr>
   <td>Kubernetes
   </td>
   <td>Cannot schedule pods
   </td>
   <td>
   </td>
   <td>Other performance issues
   </td>
  </tr>
  <tr>
   <td>Harbor
   </td>
   <td>Cannot pull images
   </td>
   <td>Cannot push images
   </td>
   <td>Performance issues
   </td>
  </tr>
  <tr>
   <td>Postgres
   </td>
   <td>
   </td>
   <td>
   </td>
   <td>No replicas/HA lost
   </td>
  </tr>
  <tr>
   <td>Vault
   </td>
   <td>
   </td>
   <td>
   </td>
   <td>
   </td>
  </tr>
  <tr>
   <td>Elasticsearch
   </td>
   <td>
   </td>
   <td>
   </td>
   <td>
   </td>
  </tr>
</table>

## Escalations

After receiving a new alert or ticket, spend a few minutes (&lt; 5m) do a preliminary investigation to identify the reproducibility and impact of the incident. Once confirmed as an issue, the formal incident response process is:

### Hypothesis Development

### Hypothesis Testing

### Mitigation

See [Generic mitigations](https://www.oreilly.com/content/generic-mitigations/).

#### :etcd: Etcd

|                  |                                                                             |
| ---------------- | --------------------------------------------------------------------------- |
| Slow Performance | :octicons-graph-16: Check disk I/O <br>:worker: Reduce size of etcd cluster |
| Loss of Quorum   | See [Disaster recovery](https://etcd.io/docs/v3.4.0/op-guide/recovery/)     |
| Key Exposure     |                                                                             |
| DB Size Exceeded |                                                                             |
|                  |                                                                             |
|                  |                                                                             |

#### :kubernetes: Kubernetes

##### Health Checks :heart:

* :bash: `karina status` (lists control plane/etcd versions/leaders/orphans)
* :bash: `karina status pods`
* :octicons-graph-16: control-plane logs [TODO - elastic query]
* :octicons-graph-16: karma, canary alerts
* :bash: [kubectl-popeye](https://github.com/derailed/popeye)

|                               |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Deployments                   | See [Troubleshooting Deployments](https://learnk8s.io/troubleshooting-deployments) <br>:bash: [kubectl-debug](https://github.com/aylei/kubectl-debug)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| No Scheduling                 | :worker: Manual schedule by specifying node name in spec                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| Network Connectivity          | See [Guide to K8S Networking](https://itnext.io/an-illustrated-guide-to-kubernetes-networking-part-1-d1ede3322727) and :material-video: [Networking Model](https://sookocheff.com/post/kubernetes/understanding-kubernetes-networking-model/) [Packet-level Debugging](https://www.youtube.com/watch?v=RQNy1PHd5_A) <br>:bash: [kubectl-sniff](https://github.com/eldadru/ksniff) – tcpdump specific pods <br/>:bash: [kubectl-tap](https://soluble-ai.github.io/kubetap/) – expose services locally<br/>:bash: [tcpprobe](https://github.com/mehrdadrad/tcpprobe) – measure 60+ metrics for socket connections <br/>:octicons-graph-16: Check node to node connectivity using :material-docker: [goldpinger](https://github.com/bloomberg/goldpinger)<br/>:worker: Restart CNI controllers/agents <br> |
| End User Access Denied        | :action: Temporarily increase access level                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| End User Access Denied        | Check access using :bash: [rbac-matrix](https://github.com/corneliusweig/rakkess)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Disk/Volume Space             | Check PV usage using  :bash: [kubectl-df-pv](https://github.com/yashbhutwala/kubectl-df-pv)<br/>:worker: Remove old filebeat/journal logs <br>:worker: Scale down replicated storage <br>:worker: Reduce replicas from 3 → 2 → 1                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| DNS Latency                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Failing Webhooks              |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Loss of Control Plane Access  |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Failure During Rolling Update | Run :bash: `karina terminate-node` followed by `karina provision`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|                               |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Node Performance              | See [Performance Cookbook](https://publib.boulder.ibm.com/httpserv/cookbook/) and [USE](http://www.brendangregg.com/USEmethod/use-linux.html)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| Unable to SSH                 | Try :bash: [kubectl-node-shell](https://github.com/kvaps/kubectl-node-shell)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |

#### :postgres: Postgresql

|                 |     |
| --------------- | --- |
| Slow Performace |     |
| Date Loss       |     |
| Key Exposure    |     |
| Failover        |     |

#### :vault: Vault / :consul: Consul

|                 |     |
| --------------- | --- |
| Slow Performace |     |
| Date Loss       |     |
| Key Exposure    |     |
| Failover        |     |

#### :harbor: Harbor

|                 |     |
| --------------- | --- |
| Slow Performace |     |
| Date Loss       |     |
| Key Exposure    |     |
| Failover        |     |

#### :elastic: Elasticsearch

|                 |     |
| --------------- | --- |
| Slow Performace |     |
| Date Loss       |     |
| Key Exposure    |     |
| Failover        |     |

## Incident Resolution
