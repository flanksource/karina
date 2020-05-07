

Configure the backup location:

```yaml
s3:
  #optional endpoint for S3-compatible blob stores
  #endpoint: 
  region: eu-east-1
  access_key: !!env AWS_ACCESS_KEY
  secret_key: !!env AWS_SECRET_KEY
velero:
  bucket: backups
```

To run a backup of the cluster objects:

```shell
platform-cli backup
```

