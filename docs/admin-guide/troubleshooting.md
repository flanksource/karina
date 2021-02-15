
## Incident Roles

|                                |                                          |
| ------------------------------ | ---------------------------------------- |
| :worker: Incident Responder(s) | The person who has the hands-on keyboard |
| :captain: Incident Commander   | Communicating with stakeholders          |
| :talk: Incident Communicator   | Overseeing the incident                  |



The steps below are applicable for Sev 1 & 2 incidents, see the Escalations section for more detailed steps with times etc..

<span style="font-size: 28px; color: grey">:1:</span> The first responder automatically becomes the incident commander until it is transferred:      <br/>

!!! warning
    You are still the :captain: commander until someone else says  :speak:  *I have command*

<span style="font-size: 28px; color: grey; ">:2:</span> The incident commander is also automatically the incident communicator until it is transferred:

!!! warning
    You are still :talk:  communicator until someone else says   :speak:  *I have comms*

<span style="font-size: 28px; color: grey">:3:</span> Setup a communications channel on %%{support.channel}%%

<span style="font-size: 28px; color: grey">:4:</span> Communicate a new incident and channel details to %%{support.email}%%



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
   <td>Cannot Pull  Images
   </td>
   <td>Cannot Push Images
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
   <td>No replicas / HA lost
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

After receiving a new alert or ticket, spend a few minutes (&lt; 5m) doing a preliminary investigation to identify the reproducibility and impact of the incident. Once confirmed as an issue being the formal incident response process:



## Hypothesis development


## Hypothesis Testing


## Mitigation

See [Generic mitigations](https://www.oreilly.com/content/generic-mitigations/)

### :etcd: Etcd

|                  |                                                              |
| ---------------- | ------------------------------------------------------------ |
| Slow Performance | :octicons-graph-16: Check disk I/O <br>:worker: Reduce size of etcd cluster |
| Loss of Quorum   | See <a href="https://etcd.io/docs/v3.4.0/op-guide/recovery/">Disaster recovery</a> |
| Key Exposure     |                                                              |
| DB Size Exceeded |                                                              |
|                  |                                                              |
|                  |                                                              |



### :kubernetes: Kubernetes

Health Checks:

* :bash: `karina status` (lists control-plane / etcd versions / leaders / orphans )
* :bash: `karina status pods`
* :octicons-graph-16: control-plane logs [TODO - elastic query]
* :octicons-graph-16: karma,canary alerts
* :bash: [kubectl-popeye](https://github.com/derailed/popeye)



|                               |                                                              |
| ----------------------------- | ------------------------------------------------------------ |
| Deployments                   | See [Troubleshooting Deployments](https://learnk8s.io/troubleshooting-deployments) <br>:bash: <a href="https://github.com/aylei/kubectl-debug">kubectl-debug</a> |
| No Scheduling                 | :worker: Manual schedule by specifying node name in spec     |
| Network Connectivity          | See <a href="https://itnext.io/an-illustrated-guide-to-kubernetes-networking-part-1-d1ede3322727">Guide to K8S Networking</a> and :material-video: <a href="https://sookocheff.com/post/kubernetes/understanding-kubernetes-networking-model/">Networking Model</a>  <a href="https://www.youtube.com/watch?v=RQNy1PHd5_A">Packet-level Debugging</a> <br>:bash: <a href="https://github.com/eldadru/ksniff">kubectl-sniff</a> - tcpdump specific pods<br/>:bash: <a href="https://soluble-ai.github.io/kubetap/">kubectl-tap</a>   - expose services locally<br/>:bash: <a href="https://github.com/mehrdadrad/tcpprobe">tcpprobe</a> - measure 60+ metrics for socket connections <br/>:octicons-graph-16: check node to node connectivity using :material-docker: <a href="https://github.com/bloomberg/goldpinger">goldpinger</a> :  <br/>:worker: Restart CNI controllers / agents <br> |
| End User Access Denied        | :action: Temporarily increase access level                   |
| End User Access Denied        | Check access using :bash: <a href="https://github.com/corneliusweig/rakkess">rbac-matrix</a> |
| Disk / Volume Space           | Check PV usage using  :bash: <a href="https://github.com/yashbhutwala/kubectl-df-pv">kubectl-df-pv</a><br/>:worker: Remove old filebeat/journal logs <br>:worker: Scale down replicated storage <br>:worker:Reduce replicas from 3 → 2 → 1 |
| DNS Latency                   |                                                              |
| Failing Webhooks              |                                                              |
| Loss of Control Plane Access  |                                                              |
| Failure during rolling update | Run :bash: `karina terminate-node` followed by `karina provision` |
|                               |                                                              |
| Node Performance              | See <a href="https://publib.boulder.ibm.com/httpserv/cookbook/">Performance Cookbook</a> and <a href="http://www.brendangregg.com/USEmethod/use-linux.html">USE</a> |
| Unable to SSH                 | Try :bash: <a href="https://github.com/kvaps/kubectl-node-shell">kubectl-node-shell</a> |



### :postgres: Postgresql

|                 |      |
| --------------- | ---- |
| Slow Performace |      |
| Date Loss       |      |
| Key Exposure    |      |
| Failover        |      |




### :vault: Vault / :consul: Consul


|                 |      |
| --------------- | ---- |
| Slow Performace |      |
| Date Loss       |      |
| Key Exposure    |      |
| Failover        |      |



### :harbor: Harbor

|                 |      |
| --------------- | ---- |
| Slow Performace |      |
| Date Loss       |      |
| Key Exposure    |      |
| Failover        |      |



### :elastic: Elasticsearch

|                 |      |
| --------------- | ---- |
| Slow Performace |      |
| Date Loss       |      |
| Key Exposure    |      |
| Failover        |      |






## Incident Resolution
