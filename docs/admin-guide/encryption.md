### Configure Kubernetes Encryption

Karina supports configuring encryption providers for stored secrets.

This is configured by supplying an encryption provider configuration file (look [here](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#encrypting-your-data) for an example).

Update the `kubernetes.encryption` section with the config file:

```yaml
kubernetes:
  encryption:
    encryptionProviderConfigFile: ./encryption-config.yaml
```

!!! warning
    Note that encryption options are only used on provisioning, to update or add auditing to an existing cluster the configuration needs to be updated and then all master nodes rolled.

See the official Kubernetes documentation for the [configuration file](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#understanding-the-encryption-at-rest-configuration) and the [various providers](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers) that can be configured.