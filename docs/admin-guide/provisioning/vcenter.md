Configure your environment with `GOVC_*` variables and then map them into the configuration file using the `!!env` YAML tag:


## Install govc
=== "Linux"
    ```bash
    wget -nv -nc https://github.com/vmware/govc/releases/latest/download/govc_linux_amd64.gz && \
        gzip -d govc_linux_amd64.gz && \
        chmod +x govc_linux_amd64 && \
        mv govc_linux_amd64 /usr/local/bin/govc
    ```

=== "MacOSX"
    ```zsh
    wget -nv -nc https://github.com/vmware/govmomi/releases/latest/download/govc_darwin_amd64.gz  && \
      gzip -d govc_darwin_amd64.gz  && \
      chmod +x govc_darwin_amd64 && \
      mv govc_darwin_amd64 /usr/local/bin/govc
    ```


## Configure govc

```bash
export GOVC_USER=
export GOVC_DATACENTER=
export GOVC_CLUSTER=
export GOVC_FOLDER=
export GOVC_DATASTORE=
# can be found on the Datastore summary page
export GOVC_DATASTORE_URL=
export GOVC_PASS=
export GOVC_FQDN=
```

## Test Connection

```bash
export GOVC_URL="$GOVC_USER:$GOVC_PASS@$GOVC_FQDN"
govc about
```
