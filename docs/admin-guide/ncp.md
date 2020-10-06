

#### Configure the required properties for the CNI

```yaml
# private docker registry
dockerRegistry:
calico:
	disabled: true
nsx:
  # T0 router id
  tier0:
  # version of container to use, the container must be published to {{dockerRegistry}}/library/nsx-ncp-ubuntu:{{version}}
  version: 2.5.1.15287458
  use_stderr: true
  # load balancer pool id
  loadbalancer_ip_pool:
  debug: false
  nsx_k8s:
    loglevel: WARNING
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    client_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
  coe:
    enable_snat: false
    loglevel: WARNING
    nsxlib_loglevel: WARNING
  nsx_v3:
    nsx_api_user: !!env NSX_USER
    nsx_api_password: !!env NSX_PASS
    policy_nsxapi: true
    top_firewall_section_marker: K8S-Top-Section
    bottom_firewall_section_marker: K8S-Bottom-Section
    use_native_loadbalancer: false
    insecure: true
    subnet_prefix: 24
    service_size: SMALL
    # TO router name
    top_tier_router:
    # Overlay TZ ID
    overlay_tz:
    # List of IP pool's for pod IP's
    no_snat_ip_blocks:
    # list of NSX API Managers
    nsx_api_managers:
  nsx_node_agent:
    ovs_bridge: br-int
    ovs_uplink_port: ens192
    log_level: WARNING

```



### Updating logging levels

To update the logging level of all the node agents and controllers, run:

```shell
karina nsx set-log-level --log-level INFO
```



