Postgres databases can be deployed using the Zalando [Postgres Operator](https://github.com/zalando/postgres-operator)

`karina.yml`

```yaml
postgresOperator:
  version: v1.3.4.flanksource.1
templateOperator:
  version: v0.1.11
canaryChecker:
  version: v0.15.1
```

Deploying using :

```bash
karina deploy postgres-operator -c karina.yml
```

 A CRD called [PostgresqlDB](https://github.com/flanksource/karina/blob/master/manifests/template/postgres-db.yaml.raw) is used as a wrapper around the core zalando objects

Once the operator is deployed you can create a new database

`db.yml`

```yaml
apiVersion: db.flanksource.com/v1
kind: PostgresqlDB
metadata:
  name: db
  namespace: postgres-operator
spec:
  backup:
    bucket: postgres-backups
    schedule: "0 */4 * * *"
  cpu: 4000m
  memory: 8Gi
  replicas: 3
  storage:
    size: 200Gi
  parameters:
    archive_mode: "on"
    archive_timeout: 60s
    log_destination: stderr
    max_connections: "600"
    shared_buffers: 2048MB

```

```bash
kubectl apply -f db.yml
```

The template operator will pick up the new `db.flanksource.com/v1` object and create underlying Zalando `Postgres` objects, `CronJobs` for backups and 2 `Canary`'s - 1 for the backup freshness and another for connecting to the postgres instance



## Day 2 Tasks

### Failover

### Clone

This command will create a new database cluster restored from WAL backup of another cluster.

```bash
karina db clone --name test1-clone --clone-cluster-name postgres-test1 --clone-timestamp "2020-04-05 14:01:00 UTC"
```

See [karina db clone](../../../cli/karina_db_clone/) documentation for all command line arguments.

### Backup

This command will perform a logical backup of the given cluster.

```
# Run backup once
karina db backup --name test1
# Deploy a cron job to run a backup every day at 04:00 AM
karina db backup --name test1 --schedule "0 4 * * *"
```

See [karina db backup](../../../cli/karina_db_backup/) documentation for all command line arguments.

### Restore

This command will restore a given cluster from a previous logical backup

```bash
karina db restore http://path/to/backup --name test1
```

See [karina db restore](../../../cli/karina_db_restore/) documentation for all command line arguments.

### Port Forwarding

1. Retrieve the password

```shell
kubectl get secret postgres.postgres-{DB-NAME}.credentials -o json -n postgres-operator | jq -r '.data.password' | base64 -D
```

2. Port forward the DB port

```shell
kubectl port-forward  po postgres-{DB-NAME}-0 5432 -n postgres-operator
```

3. Connect to the database via `localhost:5432`



