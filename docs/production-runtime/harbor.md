# Harbor

#### Backup

There are 2 components to the backup of Harbor:

1) The database is backed up on a regular basis using a Kubernetes cronjob to S3

2) The blobs themselves are not backed up, The underlying blob storage handles replication and recovery of blobs

#### Recovery

Run a database restore:

```shell
karina db restore --name postgres-harbor s3://path/to/logical_backup.tgz
```

!!! warning
      After database restore it might be required to [reset](#resetting-the-admin-password) the admin password

#### Master -> Slave Replication

(info) Should images/blobs be deleted from the primary instance and need to be recovered than a failover to the standby must take place which uses a different bucket.

#### Failover

Failover requires flipping the DNS endpoint and will invalidate existing tokens and and robot accounts potentially requiring recreation

#### Resetting the admin password
