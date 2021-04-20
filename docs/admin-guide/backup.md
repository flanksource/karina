`karina.yml`

```yaml
s3:
  #optional endpoint for S3-compatible blob stores
  #endpoint:
  region: eu-east-1
  access_key: !!env AWS_ACCESS_KEY
  secret_key: !!env AWS_SECRET_KEY
velero:
  version: v1.3.2
  bucket: backups
```
Deploy using:
```bash
karina deploy velero -c karina.yml
```

#### To run a backup of the cluster objects

```shell
karina backup
```

