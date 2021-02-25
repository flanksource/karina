
???+ asterix "Prerequisites"
    [canary-checker](/admin-guide/canary-checker) has been setup by the cluster administrator


:1: Define a canary in `canary.yml`:
```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: http-pass
spec:
  interval: 30
  http:
    - endpoint: https://httpstat.us/200
      thresholdMillis: 3000
      responseCodes: [201, 200, 301]
      responseContent: ""
      maxSSLExpiry: 7
```
:2:  Deploy the canary check using:
```bash
kubectl apply -f canary.yml
```
:3:  Check the status of the canary
```bash
NAMESPACE         NAME         INTERVAL   STATUS   MESSAGE   UPTIME 1H      LATENCY 1H   LAST TRANSITIONED   LAST CHECK
default           http-pass    30         Passed             120/120 (100%)     10ms         2d                  6s
```


## Check Types

### DNS

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: dns
spec:
  dns:
    - server: 8.8.8.8
      port: 53
      query: "flanksource.com"
      querytype: "A"
      minrecords: 1
      exactreply: ["34.65.228.161"]
      timeout: 10
```

| Field           | Description | Scheme   | Required |
| --------------- | ----------- | -------- | -------- |
| description     |             | string   | Yes      |
| server          |             | string   | Yes      |
| port            |             | int      | Yes      |
| query           |             | string   |          |
| querytype       |             | string   | Yes      |
| minrecords      |             | int      |          |
| exactreply      |             | []string |          |
| timeout         |             | int      | Yes      |
| thresholdMillis |             | int      | Yes      |


### HTTP

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: http
spec:
  http:
    - endpoint: https://httpstat.us/200
      thresholdMillis: 3000
      responseCodes: [201,200,301]
      responseContent: ""
      maxSSLExpiry: 60
    - endpoint: https://httpstat.us/500
      thresholdMillis: 3000
      responseCodes: [500]
      responseContent: ""
      maxSSLExpiry: 60
    - endpoint: https://httpstat.us/500
      thresholdMillis: 3000
      responseCodes: [302]
      responseContent: ""
      maxSSLExpiry: 60
    - namespace: k8s-https-namespace
      thresholdMillis: 3000
      responseCodes: [200]
      responseContent: ""
      maxSSLExpiry: 60

```

| Field           | Description                                                  | Scheme | Required         |
| --------------- | ------------------------------------------------------------ | ------ | ---------------- |
| description     |                                                              | string | Yes              |
| endpoint        | HTTP endpoint to monitor                                     | string | Yes <sup>*</sup> |
| namespace       | Kubernetes namespace to monitor, Specify a namespace of `"*"` to crawl all namespaces. | string | Yes <sup>*</sup> |
| thresholdMillis | Maximum duration in milliseconds for the HTTP request. It will fail the check if it takes longer. | int    | Yes              |
| responseCodes   | Expected response codes for the HTTP Request.                | []int  | Yes              |
| responseContent | Exact response content expected to be returned by the endpoint. | string | Yes              |
| maxSSLExpiry    | Maximum number of days until the SSL Certificate expires.    | int    | Yes              |

<sup>*</sup> One of either endpoint or namespace must be specified, but not both.

### Helm

Build and push a helm chart

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: ping
spec:
  helm:
    description: chart for testing
    chartmuseum: chartmuseum.harbor.svc
    username: admin
    password: admin
```



| Field       | Description | Scheme  | Required |
| ----------- | ----------- | ------- | -------- |
| description |             | string  | Yes      |
| chartmuseum |             | string  | Yes      |
| project     |             | string  |          |
| username    |             | string  | Yes      |
| password    |             | string  | Yes      |
| cafile      |             | *string |          |


### ICMP

This check will check ICMP packet loss and duration.

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: ping
spec:
  icmp:
    - endpoints:
        - https://google.com
        - https://yahoo.com
      thresholdMillis: 400
      packetLossThreshold: 0.5
      packetCount: 2
```

| Field               | Description | Scheme | Required |
| ------------------- | ----------- | ------ | -------- |
| description         |             | string | Yes      |
| endpoint            |             | string | Yes      |
| thresholdMillis     |             | int64  | Yes      |
| packetLossThreshold |             | int64  | Yes      |
| packetCount         |             | int    | Yes      |


### LDAP

The LDAP check will:

* bind using provided user/password to the ldap host. Supports ldap/ldaps protocols.
* search an object type in the provided bind DN.s

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: ldap
spec:
  ldap:
    - host: ldap://127.0.0.1:10389
      username: uid=admin,ou=system
      password: secret
      bindDN: ou=users,dc=example,dc=com
      userSearch: "(&(objectClass=organizationalPerson))"
    - host: ldap://127.0.0.1:10389
      username: uid=admin,ou=system
      password: secret
      bindDN: ou=groups,dc=example,dc=com
      userSearch: "(&(objectClass=groupOfNames))"
```

| Field         | Description | Scheme | Required |
| ------------- | ----------- | ------ | -------- |
| description   |             | string | Yes      |
| host          |             | string | Yes      |
| username      |             | string | Yes      |
| password      |             | string | Yes      |
| bindDN        |             | string | Yes      |
| userSearch    |             | string | Yes      |
| skipTLSVerify |             | bool   | Yes      |


### Postgres

This check will try to connect to a specified Postgresql database, run a query against it and verify the results.

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: psql
spec:
  postgres:
    - connection: "user=postgres password=mysecretpassword host=192.168.0.103 port=15432 dbname=postgres sslmode=disable"
      query:  "SELECT 1"
      results: 1
```

| Field       | Description | Scheme | Required |
| ----------- | ----------- | ------ | -------- |
| description |             | string | Yes      |
| driver      |             | string | Yes      |
| connection  |             | string | Yes      |
| query       |             | string | Yes      |
| results     |             | int    | Yes      |


### S3

This check will verify reachability and correctness of an S3 compatible store:

* list objects in the bucket to check for Read permissions
* PUT an object into the bucket for Write permissions
* download previous uploaded object to check for Get permissions

```yaml
apiVersion: canaries.flanksource.com/v1
kind: Canary
metadata:
  name: ldap
spec:
  s3:
    - buckets:
        - name: "test-bucket"
          region: "us-east-1"
          endpoint: "https://test-bucket.s3.us-east-1.amazonaws.com"
      secretKey: "<access-key>"
      accessKey: "<secret-key>"
      objectPath: "path/to/object"
```

| Field         | Description                           | Scheme            | Required |
| ------------- | ------------------------------------- | ----------------- | -------- |
| description   |                                       | string            | Yes      |
| bucket        |                                       | [Bucket](#bucket) | Yes      |
| accessKey     |                                       | string            | Yes      |
| secretKey     |                                       | string            | Yes      |
| objectPath    |                                       | string            | Yes      |
| skipTLSVerify | Skip TLS verify when connecting to s3 | bool              | Yes      |


### S3 Bucket

This check will query the contents of an S3 bucket for freshness

- search objects matching the provided object path pattern
- check that latest object is no older than provided MaxAge value in seconds
- check that latest object size is not smaller than provided MinSize value in bytes.

```yaml
s3Bucket:
  - bucket: foo
    accessKey: "<access-key>"
    secretKey: "<secret-key>"
    region: "us-east-2"
    endpoint: "https://s3.us-east-2.amazonaws.com"
    objectPath: "(.*)archive.zip$"
    readWrite: true
    maxAge: 5000000
    minSize: 50000
```

| Field         | Description                                                  | Scheme | Required |
| ------------- | ------------------------------------------------------------ | ------ | -------- |
| description   |                                                              | string | Yes      |
| bucket        |                                                              | string | Yes      |
| accessKey     |                                                              | string | Yes      |
| secretKey     |                                                              | string | Yes      |
| region        |                                                              | string | Yes      |
| endpoint      |                                                              | string | Yes      |
| objectPath    | glob path to restrict matches to a subset                    | string | Yes      |
| readWrite     |                                                              | bool   | Yes      |
| maxAge        | maximum allowed age of matched objects in seconds            | int64  | Yes      |
| minSize       | min size of of most recent matched object in bytes           | int64  | Yes      |
| usePathStyle  | Use path style path: http://s3.amazonaws.com/BUCKET/KEY instead of http://BUCKET.s3.amazonaws.com/KEY | bool   | Yes      |
| skipTLSVerify | Skip TLS verify when connecting to s3                        | bool   | Yes      |

