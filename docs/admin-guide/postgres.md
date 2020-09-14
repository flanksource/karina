# Databases

Postgres databases can be deployed using the [Postgres Operator](https://github.com/flanksource/postgres-operator)

### Create database

```bash
karina db create --name test1
```

See [karina db create](../../../cli/karina_db_create/) documentation for all command line arguments.

### Clone database

This command will create a new database cluster restored from WAL backup of another cluster.

```bash
karina db clone --name test1-clone --clone-cluster-name postgres-test1 --clone-timestamp "2020-04-05 14:01:00 UTC"
```

See [karina db clone](../../../cli/karina_db_clone/) documentation for all command line arguments.

### Backup database

This command will perform a logical backup of the given cluster.

```
# Run backup once
karina db backup --name test1
# Deploy a cron job to run a backup every day at 04:00 AM
karina db backup --name test1 --schedule "0 4 * * *"
```

See [karina db backup](../../../cli/karina_db_backup/) documentation for all command line arguments.

### Restore database

This command will restore a given cluster from a previous logical backup

```bash
karina db restore http://path/to/backup --name test1
```

See [karina db restore](../../../cli/karina_db_restore/) documentation for all command line arguments.

### Connect to an exsiting database

1) Retrieve the password

```shell
kubectl get secret postgres.postgres-{DB-NAME}.credentials -o json -n postgres-operator | jq -r '.data.password' | base64 -D
```

2) Port forward the DB port

```shell
kubectl port-forward  po postgres-{DB-NAME}-0 5432 -n postgres-operator
```

3) Connect to the database via `localhost:5432`



