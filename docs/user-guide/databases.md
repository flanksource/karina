Postgres databases can be deployed using the [Zalando Postgres Operator](https://github.com/flanksource/postgres-operator)



### Creating a managed postgres database

```yaml
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  namespace: postgres-operator
  name: postgres-dev-nodeport
spec:
  databases:
    pace: app
  dockerImage: docker.io/flanksource/spilo:1.6-p2.flanksource
  enableShmVolume: true
  numberOfInstances: 2
  patroni:
    initdb:
      data-checksums: "true"
      encoding: UTF8
      locale: en_US.UTF-8
    loop_wait: 10
    maximum_lag_on_failover: 33554432
    pg_hba:
      - hostssl all all 0.0.0.0/0 md5
      - host    all all 0.0.0.0/0 md5
    retry_timeout: 10
    slots: {}
    ttl: 30
  podAnnotations: {}
  postgresql:
    parameters:
      archive_mode: "on"
      archive_timeout: 3600s
      max_connections: "200"
    version: "12"
  resources:
    limits:
      cpu: "2"
      memory: 2Gi
    requests:
      cpu: "1"
      memory: 1Gi
  serviceAnnotations: {}
  teamId: postgres
  users:
    app:
      - createdb
      - superuser
  volume:
    size: 20Gi
    storageClass: vsan
```



### Setting up scheduled backups



```yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: backup-postgres-YOUR-APP
  namespace: postgres-operator
spec:
  concurrencyPolicy: Forbid
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            application: spilo-logical-backup
            cluster-name: postgres-YOUR-APP
        spec:
          containers:
            - env:
                - name: POD_NAMESPACE
                  valueFrom:
                    fieldRef:
                      apiVersion: v1
                      fieldPath: metadata.namespace
                - name: PGPASSWORD
                  valueFrom:
                    secretKeyRef:
                      key: password
                      name: postgres.postgres-app-dev.credentials
                - name: LOGICAL_BACKUP_S3_ENDPOINT
                  value: https://YOUR_S3_ENDPOINT
                - name: PGHOST
                  value: postgres-app-dev
                - name: AWS_SECRET_ACCESS_KEY
                  valueFrom:
                    secretKeyRef:
                      key: AWS_SECRET_ACCESS_KEY
                      name: secrets
                - name: PGPORT
                  value: "5432"
                - name: LOGICAL_BACKUP_S3_REGION
                  value: 1dp
                - name: AWS_ACCESS_KEY_ID
                  value: YOUR_AWS_ACCESS_KEY
                - name: SCOPE
                  value: postgres-YOUR-APP
                - name: PGSSLMODE
                  value: prefer
                - name: PGDATABASE
                  value: postgres
                - name: LOGICAL_BACKUP_S3_BUCKET
                  value: YOUR_S3_BACKUP_BUCKET
                - name: LOGICAL_BACKUP_S3_SSE
                  value: AES256
                - name: PG_VERSION
                  value: "12"
                - name: PGUSER
                  value: postgres
                - name: CLUSTER_NAME_LABEL
                  value: cluster-name
              image: docker.io/flanksource/postgres-backups:0.1.5
              imagePullPolicy: IfNotPresent
              name: backup-postgres-YOUR-APP
              ports:
                - containerPort: 8080
                  protocol: TCP
                - containerPort: 5432
                  protocol: TCP
                - containerPort: 8008
                  protocol: TCP
              resources:
                limits:
                  cpu: 500m
                  memory: 512Mi
                requests:
                  cpu: 10m
                  memory: 128Mi
          securityContext: {}
          serviceAccount: postgres-pod
          serviceAccountName: postgres-pod
          terminationGracePeriodSeconds: 30
  schedule: "@midnight"
  successfulJobsHistoryLimit: 3
  suspend: false
```


```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-dev-nodeport
  namespace: postgres-operator
spec:
  ports:
    - name: postgresql
      port: 5432
      protocol: TCP
      targetPort: 5432
      nodePort: 30202
  type: NodePort
  selector:
    cluster-name: postgres-dev-nodeport
    spilo-role: master
```

