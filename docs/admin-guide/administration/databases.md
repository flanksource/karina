# Databases

Postgres databases can be deployed using the [Postgres Operator](https://github.com/flanksource/postgres-operator)

### Create database

```bash
platform-cli db create --name test1
```

See [platform-cli db create](../../../cli/platform-cli_db_create/) documentation for all command line arguments.

### Clone database

This command will create a new database cluster restored from WAL backup of another cluster.

```bash
platform-cli db clone --name test1-clone --clone-cluster-name postgres-test1 --clone-timestamp "2020-04-05 14:01:00 UTC" 
```

See [platform-cli db clone](../../../cli/platform-cli_db_clone/) documentation for all command line arguments.

### Backup database

This command will perform a logical backup of the given cluster.

```
# Run backup once
platform-cli db backup --name test1
# Deploy a cron job to run a backup every day at 04:00 AM
platform-cli db backup --name test1 --schedule "0 4 * * *" 
```

See [platform-cli db backup](../../../cli/platform-cli_db_backup/) documentation for all command line arguments.

### Restore database

This command will restore a given cluster from a previous logical backup

```bash
platform-cli db restore http://path/to/backup --name test1
```

See [platform-cli db restore](../../../cli/platform-cli_db_restore/) documentation for all command line arguments.

