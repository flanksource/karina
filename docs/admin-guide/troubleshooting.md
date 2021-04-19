
## Incident Roles

|                                |                                 |
| ------------------------------ | ------------------------------- |
| :worker: Incident Responder(s) | Hands on keyboard               |
| :captain: Incident Commander   | Communicating with stakeholders |
| :talk: Incident Communicator   | Overseeing the incident         |

The steps below are applicable for Severity ("**Sev**") 1 & 2 level incidents. See the [Escalations](#escalations) [section][foo] for more detailed steps and [Severity Classification](#severity-classification) for a severity classification rubric.

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
   <td>Performance issues/Cannot use UI
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

After receiving a new alert or ticket, spend a few minutes (&lt; 5m) doing a preliminary investigation to identify the reproducibility and impact of the incident. Once confirmed as an issue, the formal incident response process is:

### Observing the environment

:1: Start by checking general cluster health

:2: Check recent changes that may have had an impact

:3: Check recent trends in consumption metrics:

* Has a threshold been reached? A threshold can be both fixed (e.g. memory) and cumulative (disk space).

:4: Check performance metrics

* Are metrics within bounds?
* Are any metrics abnormal?

:5:

### Hypothesis Development & Testing

- What do I think is the problem?
- What would it look like if I were right about what the problem is?
- Could anything other problem present in this way?
- Of all the possible causes I've identified, are any much more likely than others?
- Of all the possible causes I've identified, are any much more costly to investigate and fix than others?
- How can I prove that, of all the problems that present in this way, the problem that I think is the problem is in fact the problem?
- How can I test my guesses in a way that, if I'm wrong, I still have a system that is meaningfully similar to the one that I was presented with at the start?

Why even talk about 'hypotheses' rather than 'ideas' or guesses'? Using more formal scientific language is a choice that signals a connection to – and that creates a bridge for – a larger set of ideas and techniques about evidence and progress in knowledge. Because of the caché scientific ideas hold, it's important to understand that much of theory is constructed in hindsight – science, like engineering, is a messy, human business and there is no perfect method for developing and testing hypotheses. Whatever is included here will necessarily leave much out, and there will always be cases that are not covered or in which the advice below misleads – nevertheless, it has been our experience that the following set of questions are a useful [heuristic](https://en.wikipedia.org/wiki/Heuristic).

!!! info "Interest readings"
    - [Wikipedia: Hypothetico-deductive model](https://en.wikipedia.org/wiki/Hypothetico-deductive_model)
    - [Wikipedia: Differential diagnosis – Specific_methods](https://en.wikipedia.org/wiki/Differential_diagnosis#Specific_methods)
    - [Google SRE Handbook: Effective Troubleshooting – Test and Treat](https://sre.google/sre-book/effective-troubleshooting/#test-and-treat-EvsWceuj)

#### What do I think is the cause? – Forming the conjecture

A hypothesis, or a conjecture, is a guess about what causes some feature of the world. A useful hypothesis can be tested – ie, there is some way to say whether it is true or not. It's a very ordinary thing that engineers do everyday during most hours of the day.

An engineer's capacity to form good, testable hypotheses about a cluster or piece of software is a function of their capacity, experience and observation tooling – knowledge, as they say, is power. When the cause of a problem is unknown, surveying the appropriate data is a crucial first step in developing useful and clear hypotheses about it. See [observing the environment](#observing-the-environment) for advice on gathering the relevant information.

Let's start with a problem/symptom:

!!! example "The Symptom"
    All endpoints report as down on the monitoring tool.

Then we generate a hypothesis about what might be the cause:

!!! example "Hypothesis 1.1"
    All the cluster pods are crashing.

#### What would it look like if I were right about what the problem is? – Making testable predictions

Useful hypotheses are testable, which means that if they are true, the world is one observable way, and if they are false, the world is another observable way. Hypothesis 1.1 is a guess about the state of the world, but it doesn't include any tests. The testable claim (Proposition 1.1) below includes a statement about what would hold were hypothesis above 1.1 true:

!!! example "Proposition 1.1"
    If all the cluster pods were crashing, I would not be able to curl any of the endpoints.

Let's stipulate for the example that there is an attempt to curl all the endpoints and all attempts fail to produce a response. What can we conclude from this?

Unfortunately, unless the **only possible cause** of being unable to curl the endpoints is that all the pods were down, we have not discovered the root cause of the symptom – we've merely noticed that the world is consistent with our hypothesis and, while it remains possible that all the pods are crashing, we've not ruled out other hypotheses that are also consistent with what we've seen. If there are any such hypotheses, we say that our conclusion is [underdetermined](https://en.wikipedia.org/wiki/Underdetermination).

Here are some others:

!!! example "Hypotheses 2.1"
    - cluster DNS is misconfigured  
    - services are misconfigured and no longer point correctly to pods
    - a local firewall is blocking the cluster IP

???+ info "Reasoning from inconsistency"
    Had we been able to curl any of the endpoints, we could infer that:

    - some of the pods have problems that prevent curling their endpoints but not all
    - we're not generally blocked by a firewall from hitting the cluster domain

    Therefore, the following might hold:

    !!! example "Hypotheses from inconsistency"
        - there are problems with the monitoring tool
        - there are problems with some service configurations
        - some pods are crashing

#### Could any other problems present in this way? – Avoiding underdetermination

A conclusion is [underdetermined](https://en.wikipedia.org/wiki/Underdetermination) by data when plausible rival hypotheses are reasonably likely to be true. For example, even if Proposition 1.1 ("If all the cluster pods were crashing, I would not be able to curl any of the endpoints.") were true and attempts to curl all the endpoints failed in each case, the following hypotheses are rival to Hypothesis 1.1 ("All the cluster pods are crashing"):

!!! example "Hypotheses 3.1"
    - a firewall is blocking access
    - cluster DNS is misconfigured
    - all service configuration has been altered
    - the cluster was deleted by an angry ex-employee
    - there has been an earthquake and the data centre and backup data centres have been destroyed

Generally speaking, we want to run tests that eliminate competing hypotheses as well as tests that affirm specific hypotheses. Many symptoms/problems will, at start, have vast numbers of plausible competing hypotheses that could be tested. Tests that properly eliminate possibilities can therefore be extremely valuable, especially if those possibilities are both reasonably likely and cheap to check.

#### Of all the possible causes I've identified, are any much more likely than others? – Assessing the base rate
#### Of all the possible causes I've identified, are any much more costly (time or money) to investigate and fix than others? – Assessing costs of investigation
#### How can I prove that, of all the problems that present in this way, the problem that I think is the problem is in fact the problem? – Ruling out competing hypotheses
#### How can I test my guesses in a way that, if I'm wrong, I still have a system that is meaningfully similar to the one that I was presented with at the start? – Preserving test–retest validity

Once you've developed a [hypothesis](#hypothesis-development) for what has gone wrong, you need to validate whether your guess is correct.

!!! warning
    In a production environment, it is essential to preserve what might be called (with debt to statistics) **test-retest reliability**, which can be defined as the reliability of a test measured over time.  

    In the case of a production cluster, configuration changes or changes to the underlying systems on which a cluster depends risk undermining test-retest reliability. By contrast, a staging cluster or local cluster is much more likely to preserve test-retest reliability, as a) it's generally possible to redeploy for testing and b) there is unlikely to be an accumulation of state. 
    
    The following strategies can be used to maximise test-retest reliability (to the degree possible and using your considered judgement):

    - Be prepared to rollback any change to initial state and do so after validating/disconfirming any hypothesis.
    - Until a solution is realised, ensure all changes are rolled back before pursuing a new idea – hypothesis testing should all be done from the from the same base configuration.
    - Log changes and rollbacks in the live incident log (see [SRE Handbook: Effective Troubleshooting/Negative Results Are Magic](https://sre.google/sre-book/effective-troubleshooting/#xref_troubleshooting_negative-results)).

### Incident log

TODO

### Mitigation

See [Generic mitigations](https://www.oreilly.com/content/generic-mitigations/).

#### :kubernetes: Kubernetes

##### Health Checks :heart:

* :bash: `karina status` (lists control plane/etcd versions/leaders/orphans)
* :bash: `karina status pods`
* :octicons-graph-16: control-plane logs [TODO - elastic query]
* :octicons-graph-16: karma, canary alerts
* :bash: [kubectl-popeye](https://github.com/derailed/popeye)

??? info "Deployments"
     See [Troubleshooting Deployments][troubleshooting-deployments]  
     :bash: [kubectl-debug]

??? info "No Scheduling "
     :worker: Manual schedule by specifying node name in spec

??? info "Network Connectivity"
    See [Guide to K8S Networking][guide-to-k8s-networking] and :material-video: [Networking Model][networking-model] [Packet-level Debugging][packet-level-debugging]  
    :bash: [kubectl-sniff][kubectl-sniff] – tcpdump specific pods  
    :bash: [kubectl-tap][kubectl-tap] – expose services locally  
    :bash: [tcpprobe][tcpprobe] – measure 60+ metrics for socket connections  
    :octicons-graph-16: Check node to node connectivity using :material-docker: [goldpinger][goldpinger]  
     :worker: Restart CNI controllers/agents  

??? info "End User Access Denied"
     :action: Temporarily increase access level
     Check access using :bash: [rbac-matrix][rbac-matrix]

??? info "Disk/Volume Space"
    Check PV usage using
    :bash: [kubectl-df-pv][kubectl-df-pv]  
    :worker: Remove old filebeat/journal logs  
    :worker: Scale down replicated storage  
    :worker: Reduce replicas from 3 → 2 → 1  

??? info "DNS Latency"
    :octicons-graph-16: Check DNS request/failure count on grafana
    :octicons-graph-16: Check pod ndots configuration and reduce if possible
    :octicons-graph-16: Check node-local cache hit rates
    :worker: Scale coredns replicas and/or increase cpu limits on coredns and node-local-dns

??? info "Failing Webhooks"
    :worker: Temporarily disable the webhooks by either deleting them or setting the **FailurePolicy** to *ignore*

??? info "Loss of Control Plane Access"
    :worker: Try gain access to master nodes and regen certs using `/etc/kubernetes/pki`  
    :worker:  Downscale cluster to 1 master, and regen certs using kubeadm, scale masters back up  

??? info "Failure During Rolling Update"
    Run :bash: `karina terminate-node` followed by `karina provision`

??? info "Worker Node failure"
    Run :bash: `karina terminate-node` followed by `karina provision`

??? info "Control Plane Node Failure"
    :action: Run :bash: `kubectl terminate-node` followed by `karina provision`  
    :worker:  Remove any failed etcd members using :bash: `karina etcd remove-member`

??? info "Namespace Overutilisation"
    TBD

??? info "Load Balancer Failure"
    TBD

??? info "Cluster Failure"
    :action: Cordon the cluster by removing GSLB entries  
    :action: Without PVCs: Run :bash: `kubectl terminate` followed by `karina provision` to reprovision the cluster  
    :action: With PVCs: Try take a backup first (`karina backup or velero backup`), provision a new cluster with a new name, restore from backup `karina restore or velero restore`

??? info "Cluster over-utilized"
    :action: Increase capacity if possible (even temporarily)  
    :action: Shed load starting with:

      * moving workloads to other clusters  
      * reducing replica counts  
      * terminating non critical workloads (dynamic namespaces, etc)

[troubleshooting-deployments]: https://learnk8s.io/troubleshooting-deployments
[kubectl-debug]: https://github.com/aylei/kubectl-debug
[packet-level-debugging]: https://www.youtube.com/watch?v=RQNy1PHd5_A
[networking-model]: https://sookocheff.com/post/kubernetes/understanding-kubernetes-networking-model/
[guide-to-k8s-networking]: https://itnext.io/an-illustrated-guide-to-kubernetes-networking-part-1-d1ede3322727
[rbac-matrix]: https://github.com/corneliusweig/rakkess
[kubectl-df-pv]: https://github.com/yashbhutwala/kubectl-df-pv
[kubectl-node-shell]: https://github.com/kvaps/kubectl-node-shell

#### Node

##### Health Checks :heart:

* Check CNI health (nsx-node-agent etc)
* Check performance
* Check network connectivity
* :octicons-graph-16: Check  karma, canary alerts
* :worker: Review journalbeat logs/log counts by node

??? info "Node Performance"
    :octicons-graph-16: Check container top using :bash: `crictl stats` and `crictl ps<br /`  
    :octicons-graph-16: Check VM host CPU/Memory/IO saturation  
    See [Performance Cookbook][performance-cookbook] and [USE][use]

??? info "Unable to SSH"
    Try :bash: [kubectl-node-shell][kubectl-node-shell]

??? info "CNI Failure"
    :worker: Run :bash: `karina terminate-node` followed by `karina provision`  

??? info "Network Connectivity"
    :bash: [kubectl-sniff][kubectl-sniff] - tcpdump specific pods  
    :bash: [kubectl-tap][kubectl-tap] – expose services locally  
    :bash: [tcpprobe][tcpprobe] – measure 60+ metrics for socket connections  
    :octicons-graph-16: check node to node connectivity using :material-docker: [goldpinger][goldpinger]

[performance-cookbook]: https://publib.boulder.ibm.com/httpserv/cookbook/
[use]: http://www.brendangregg.com/USEmethod/use-linux.html

<!--  TODO: Structure commented items to fit

Using  bpf tools:

https://github.com/iovisor/bcc#tools

bpftrace or via kubectl trace

https://packetlife.net/media/library/12/tcpdump.pdf

 ps -axfo pid,ppid,uname,cmd

`strace -fttTyyy -s 1024 -o <FILE>` and then

https://gitlab.com/gitlab-com/support/toolbox/strace-parser

-->

#### :etcd: Etcd

??? info "Slow Performance"
    :octicons-graph-16: Check disk I/O  
    :worker: Reduce size of etcd cluster

??? info "Loss of Quorum"
    See [Disaster recovery][disaster-recovery]

??? info "Key Exposure"
    TBD

??? info "DB Size Exceeded"
    :action: Run etcd compaction and/or reduce version history retention

[disaster-recovery]: https://etcd.io/docs/v3.4.0/op-guide/recovery/

#### :postgres: Postgresql

##### Health Checks :heart:

* Check replication status :bash: `kubectl exec -it postgres-<db-name>-0 -- patronictl list`
* Check volume usage :bash: `kubectl krew install df-pv; kubectl df-pv -n postgres-operator`

??? info "Disk Space Usage"
    :worker: Check size `pg_wal directory`; if it is taking up more than 10% of space then the WALs are not getting archived

??? info "Slow Performance"
    TBD

??? info "Data Loss"
    Recover from backup

??? info "Key Exposure"
    TBD

??? info "Not healthy enough for leader race"
    Replicas are too out of sync to promote; restore from backup

??? info "Replica not following"
    :bash: `kubectl exec -it postgres-<db-name>-0 -- patornictl reinit postgres-<db-name> pod-name`

??? info "Failover (Safe)"
    :bash: run `kubectl exec -it postgres-<db-name>-0 -- patronictl failover`  
    Delete config endpoint if failover is stuck without any master

??? info "Failover (Forced)"
    :bash: run `kubectl delete endpoints postgres-<db-name> postgres-<db-name>-config` to force re-election

??? info "WAL Logs not getting archived"
    :worker: Check standby that are offline  
    :worker: Cleanup manually using [pg_archivecleanup][pg_archivecleanup] :warning: Note cleaning up WAL logs will prevent standbys that are not up to date from catching up – they will need to be re-bootstrapped

[pg_archivecleanup]: https://www.percona.com/blog/2019/07/10/wal-retention-and-clean-up-pg_archivecleanup/

#### :vault: Vault / :consul: Consul

??? info "Slow Performance"
    TBD

??? info "Data Loss"
    TBD

??? info "Key Exposure"
    TBD

??? info "Failover"
    TBD

#### :harbor: Harbor

??? info "Crashlooping"
    :worker: Check health via `/api/v2.0/health`  
    `failed to migrate: please upgrade to version x first` --> :sql: `DELETE from schema_migrations where VERSION = 1`

??? info "Inaccessible"
    :worker: compare accessibility via UI/API/ `docker login`  
    :worker: Reset CSRF_TOKEN in harbor-core configmap  
    :worker: Reset admin password in `harbor_users` postgres table  
    :worker: Drop and recreate Redis PVC to flush all caches  
    :worker: Delete all running replication jobs via UI or via API

??? info "Slow Performance"
    :octicons-graph-16: Check performance of underlying registry storage (S3/disk etc)  
    :octicons-graph-16: Check CPU load/throttling on the postgres DB

??? info "Data Loss"
    :worker: Fail forward and ask dev-teams to rebuild and repush images  
    :worker: Recover images from running nodes

??? info "Key Exposure"
    TBD

??? info "Failover"
    TBD

#### :elastic: Elasticsearch

??? info "Slow Performance"
    :octicons-graph-16: Check performance of underlying disks  
    :octicons-graph-16: Check CPU load/throttling on the elastic instance  
    :octicons-graph-16: Check memory saturation under elastic node health

??? info "Data Loss"
    TBD

??? info "Key Exposure"
    TBD

??? info "Failover"
    TBD

## Incident Resolution

<!-- Section footlinks -->
[goldpinger]: https://github.com/bloomberg/goldpinger
[kubectl-tap]: https://soluble-ai.github.io/kubetap/
[kubectl-sniff]: https://github.com/eldadru/ksniff
[tcpprobe]: https://github.com/mehrdadrad/tcpprobe
