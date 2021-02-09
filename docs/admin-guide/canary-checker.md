## Deploying the operator

`karina.yml`

```yaml
domain: DOMAIN
canaryChecker:
  version: v0.15.1
```
Then deploy using:
```bash
karina deploy canary-checker -c karina.yml
```

The Canary Checker UI will be available on [https://canaries.DOMAIN]()

See the [user-guide](../user-guide/canary-checker.md) for how to configure and interact with canaries.


### Aggregation
Aggregation allows the results from multiple canary-checker instances to be view in a single page. To enable just add the endpoints for existing canary-checker instances:

```yaml
  aggregateServers:
    - https://canaries.CLUSTER1
    - https://canaries.CLUSTER2
```





## Infrastructure Checks

### Containerd

This check will try to pull a Docker image from specified registry using containers and then verify it's checksum and size.

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: containerd
spec:
  containerdPull:
    - image: docker.io/library/busybox:1.31.1
      username:
      password:
      expectedDigest: 6915be4043561d64e0ab0f8f098dc2ac48e077fe23f488ac24b665166898115a
      expectedSize: 1219782
```

| Field          | Description | Scheme | Required |
| -------------- | ----------- | ------ | -------- |
| description    |             | string | Yes      |
| image          |             | string | Yes      |
| username       |             | string | Yes      |
| password       |             | string | Yes      |
| expectedDigest |             | string | Yes      |
| expectedSize   |             | int64  | Yes      |

### Docker Pull

This check will try to pull a Docker image from specified registry, verify it's checksum and size.

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: docker
spec:
  docker:
    - image: docker.io/library/busybox:1.31.1
      username:
      password:
      expectedDigest: 6915be4043561d64e0ab0f8f098dc2ac48e077fe23f488ac24b665166898115a
      expectedSize: 1219782
```

| Field          | Description | Scheme | Required |
| -------------- | ----------- | ------ | -------- |
| description    |             | string | Yes      |
| image          |             | string | Yes      |
| username       |             | string | Yes      |
| password       |             | string | Yes      |
| expectedDigest |             | string | Yes      |
| expectedSize   |             | int64  | Yes      |


### Docker Push

| Field       | Description | Scheme | Required |
| ----------- | ----------- | ------ | -------- |
| description |             | string | Yes      |
| image       |             | string | Yes      |
| username    |             | string | Yes      |
| password    |             | string | Yes      |

### Pod

Create a new pod and verify reachability through an ingress

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: pod
spec:
  pod:
    - name: golang
      namespace: default
      spec: |
        apiVersion: v1
        kind: Pod
        metadata:
          name: hello-world-golang
          namespace: default
          labels:
            app: hello-world-golang
        spec:
          containers:
            - name: hello
              image: quay.io/toni0/hello-webserver-golang:latest
      port: 8080
      path: /foo/bar
      ingressName: hello-world-golang
      ingressHost: "hello-world-golang.127.0.0.1.nip.io"
      scheduleTimeout: 2000
      readyTimeout: 5000
      httpTimeout: 2000
      deleteTimeout: 12000
      ingressTimeout: 5000
      deadline: 29000
      httpRetryInterval: 200
      expectedContent: bar
      expectedHttpStatuses: [200, 201, 202]
```

| Field                | Description | Scheme | Required |
| -------------------- | ----------- | ------ | -------- |
| description          |             | string | Yes      |
| name                 |             | string | Yes      |
| namespace            |             | string | Yes      |
| spec                 |             | string | Yes      |
| scheduleTimeout      |             | int64  | Yes      |
| readyTimeout         |             | int64  | Yes      |
| httpTimeout          |             | int64  | Yes      |
| deleteTimeout        |             | int64  | Yes      |
| ingressTimeout       |             | int64  | Yes      |
| httpRetryInterval    |             | int64  | Yes      |
| deadline             |             | int64  | Yes      |
| port                 |             | int64  | Yes      |
| path                 |             | string | Yes      |
| ingressName          |             | string | Yes      |
| ingressHost          |             | string | Yes      |
| expectedContent      |             | string | Yes      |
| expectedHttpStatuses |             | []int  | Yes      |
| priorityClass        |             | string | Yes      |

### Namespace

The Namespace check will:

* create a new namespace using the labels/annotations provided

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: namespace
spec:
  namespace:
    - namePrefix: "test-name-prefix-"
      labels:
        team: test
      annotations:
        "foo.baz.com/foo": "bar"
```

| Field                | Description | Scheme            | Required |
| -------------------- | ----------- | ----------------- | -------- |
| description          |             | string            | Yes      |
| checkName            |             | string            | Yes      |
| namespaceNamePrefix  |             | string            | Yes      |
| namespaceLabels      |             | map[string]string | Yes      |
| namespaceAnnotations |             | map[string]string | Yes      |
| podSpec              |             | string            | Yes      |
| scheduleTimeout      |             | int64             | Yes      |
| readyTimeout         |             | int64             | Yes      |
| httpTimeout          |             | int64             | Yes      |
| deleteTimeout        |             | int64             | Yes      |
| ingressTimeout       |             | int64             | Yes      |
| httpRetryInterval    |             | int64             | Yes      |
| deadline             |             | int64             | Yes      |
| port                 |             | int64             | Yes      |
| path                 |             | string            | Yes      |
| ingressName          |             | string            | Yes      |
| ingressHost          |             | string            | Yes      |
| expectedContent      |             | string            | Yes      |
| expectedHttpStatuses |             | []int64           | Yes      |
| priorityClass        |             | string            | Yes      |
