https://thanos.DOMAIN

Thanos provides long-term persistence via S3 with multi-cluster visibility, it can be deploy in 2 modes:

#### Client Mode

In client mode thanos only persists metrics to S3 and exposes an endpoint for other Thanos query instances to query, in client mode only data older than 2h and queried from the underlung S3 store will be available for other clusters.

`karina.yml`

```yaml
monitoring:
  prometheus:
    version: v2.20.0
thanos:
  bucket: thanos-bucket # a shared S3 bucket name across all clusters
  version: v0.14.0
  mode: client
```

### Observability Mode

In observability mode, Thanos queries S3 for long-term metrics, but also queries each individual thanos client for both a long-term and real-time view of metrics.

`karina.yml`

```yaml
monitoring:
  prometheus:
    version: v2.20.0
thanos:
  bucket: thanos-bucket # a shared S3 bucket name across all clusters
  version: v0.14.0
  mode: observability
  enableCompactor: true # the compactor should only be enabled on a single cluster
  clientSidecars:
    - thanos-sidecar.cluster01.k8s:31901
    - thanos-sidecar.cluster02.k8s:31901
```

!!! warning
   The thanos sidecards need to be exported via GRPCS with TLS passthrough, there is an open issue [#1507](https://github.com/thanos-io/thanos/issues/1507) with the workaround that we apply is a adding a NodePort on **31901** to talk directory to thanos sidecards

See [thanos config options](/reference/config/#thanos)

