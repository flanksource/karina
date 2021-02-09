Karina requires a service discovery mechanism to facilitate the initial connection to the kubernetes hosts. A containerised consul service discovery can be enabled on a host in the vsphere cluster using the konfigadm tool:


# Load Balancer

# DNS

# Consul

## Create consul.yml
`consul.yml`
```yaml
commands:
  - mkdir -p /opt/consul
  - chown -R 100:1000 /opt/consul
  - iptables -A PREROUTING -t nat -p tcp --dport 80 -j REDIRECT --to-ports 8500
container_runtime:
  type: docker
containers:
  - image: docker.io/consul:1.9.1
    docker_opts: --net=host
    args: agent -server -ui -data-dir /opt/consul -datacenter lab -bootstrap
    volumes:
      - /opt/consul:/opt/consul
    env:
      CONSUL_BIND_INTERFACE: ens160
      CONSUL_CLIENT_INTERFACE: ens160
```

## Deploy consul
```bash
karina provision vm -c karina.yml -k consul.yaml`
```

