Access harbor at: [:octicons-link-external-24: consul.%%{domain}%%](https://consul.%%{domain}%%)

# Consul

### Backup

This command will create a consul snapshot and upload it to the S3 bucket provided in config.

```bash
# Run snapshot one time
karina consul backup --name consul-server --namespace vault
# Deploy a cron job to create a snapshot every day at 04:00 AM
karina consul backup --name consul-server --namespace vault --schedule "0 4 * * *"
```

See [karina consul backup](../../../cli/karina_consul_backup/) documentation for all command line arguments.


### Restore

This command will restore a consul cluster from a snapshot stored in S3.

```bash
karina consul restore --name consul-server --namespace vault s3://consul-backups/consul/backups/vault/consul-server/2020-04-03_01:02:03.snapshot
```

See [karina consul restore](../../../cli/karina_consul_restore/) documentation for all command line arguments.

