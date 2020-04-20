# Consul

### Backup

This command will create a consul snapshot and upload it to the S3 bucket provided in config.

```bash
# Run snapshot one time
platform-cli consul backup --name consul-server --namespace vault
# Deploy a cron job to create a snapshot every day at 04:00 AM
platform-cli consul backup --name consul-server --namespace vault --schedule "0 4 * * *" 
```

See [platform-cli consul backup](../../../cli/platform-cli_consul_backup/) documentation for all command line arguments.


### Restore

This command will restore a consul cluster from a snapshot stored in S3.

```bash
platform-cli consul restore --name consul-server --namespace vault s3://consul-backups/consul/backups/vault/consul-server/2020-04-03_01:02:03.snapshot
```

See [platform-cli consul restore](../../../cli/platform-cli_consul_restore/) documentation for all command line arguments.

