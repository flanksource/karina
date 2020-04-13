# Databases

Postgres databases can be deployed using the [Postgres Operator](https://github.com/flanksource/postgres-operator)

### Create database

```bash
platform-cli db create --name test1
```

```
Usage:
  platform-cli db create [flags]

Flags:
  -h, --help                     help for create
      --wal-archiving            Enable wal archiving (default true)
      --wal-enable-cluster-uid   Enable cluster UID in wal logs s3 path
      --wal-schedule string      A cron schedule to backup wal logs (default "*/5 * * * *")
      --wal-use-walg-restore     Enable wal-g for wal restore (default true)

Global Flags:
  -c, --config stringArray   Path to config file
      --dry-run              Don't apply any changes, print what would have been done
  -e, --extra stringArray    Extra arguments to apply e.g. -e ldap.domain=example.com
  -v, --loglevel count       Increase logging level
      --name string          Name of the postgres cluster / service
      --namespace string      (default "postgres-operator")
      --secret string        Name of the secret that contains the postgres user credentials
      --superuser string     Superuser user (default "postgres")
      --trace                Print out generated specs and configs
```

### Clone database

This command will create a new database cluster restored from WAL backup of another cluster.

```bash
platform-cli db clone --name test1-clone --clone-cluster-name postgres-test1 --clone-timestamp "2020-04-05 14:01:00 UTC" 
```

```
Usage:
  platform-cli db clone [flags]

Flags:
      --clone-cluster-name string   Name of the cluster to clone
      --clone-cluster-uid string    UID of the cluster to clone
      --clone-timestamp string      Timestamp of the wal to clone
  -h, --help                        help for clone
      --wal-archiving               Enable wal archiving (default true)
      --wal-enable-cluster-uid      Enable cluster UID in wal logs s3 path
      --wal-schedule string         A cron schedule to backup wal logs (default "*/5 * * * *")
      --wal-use-walg-restore        Enable wal-g for wal restore (default true)

Global Flags:
  -c, --config stringArray   Path to config file
      --dry-run              Don't apply any changes, print what would have been done
  -e, --extra stringArray    Extra arguments to apply e.g. -e ldap.domain=example.com
  -v, --loglevel count       Increase logging level
      --name string          Name of the postgres cluster / service
      --namespace string      (default "postgres-operator")
      --secret string        Name of the secret that contains the postgres user credentials
      --superuser string     Superuser user (default "postgres")
      --trace                Print out generated specs and configs
```

### Backup database

This command will perform a logical backup of the given cluster.

```
# Run backup once
platform-cli db backup --name test1
# Deploy a cron job to run a backup every day at 04:00 AM
platform-cli db backup --name test1 --schedule "0 4 * * *" 
```

```
Create a new database backup

Usage:
  platform-cli db backup [flags]

Flags:
  -h, --help              help for backup
      --schedule string   A cron schedule to backup on a reoccuring basis

Global Flags:
  -c, --config stringArray   Path to config file
      --dry-run              Don't apply any changes, print what would have been done
  -e, --extra stringArray    Extra arguments to apply e.g. -e ldap.domain=example.com
  -v, --loglevel count       Increase logging level
      --name string          Name of the postgres cluster / service
      --namespace string      (default "postgres-operator")
      --secret string        Name of the secret that contains the postgres user credentials
      --superuser string     Superuser user (default "postgres")
      --trace                Print out generated specs and configs
```

### Restore database

This command will restore a given cluster from a previous logical backup

```bash
platform-cli db restore http://path/to/backup --name test1
```

```
Restore a database from backups

Usage:
  platform-cli db restore [backup path] [flags]

Flags:
  -h, --help   help for restore

Global Flags:
  -c, --config stringArray   Path to config file
      --dry-run              Don't apply any changes, print what would have been done
  -e, --extra stringArray    Extra arguments to apply e.g. -e ldap.domain=example.com
  -v, --loglevel count       Increase logging level
      --name string          Name of the postgres cluster / service
      --namespace string      (default "postgres-operator")
      --secret string        Name of the secret that contains the postgres user credentials
      --superuser string     Superuser user (default "postgres")
      --trace                Print out generated specs and configs
```