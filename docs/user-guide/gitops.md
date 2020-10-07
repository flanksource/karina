Karina uses the [FluxCD](https://docs.fluxcd.io) controller for GitOps

![](https://docs.fluxcd.io/en/1.19.0/_files/flux-cd-diagram.png)

## Getting Started

First, Flux needs to deployed using the following karina snippet:

```yaml
gitops:
  - gitUrl: !!template https://{{ getenv "GITHUB_BOT_ACCESS_TOKEN" }}@github.com/acme-port/gitops.git
    fluxVersion: 1.20.0
```

```shell
karina deploy gitops
```

!!! warning
    If a namespace is not specified a cluster-wide flux controller is deployed into kube-system with cluster-admin privileges, ensure that the git repository it is linked to has appropriate access control

Once deployed Flux will periodically poll the Git repository and apply the matches files into the corresponding namespaces.

#### Reference

| Field           | Description                                                  | Scheme            | Required |
| --------------- | ------------------------------------------------------------ | ----------------- | -------- |
| name            | The name of the gitops deployment, defaults to namespace name | string            |          |
| disableScanning | Do not scan container image registries to fill in the registry cache, implies `--git-read-only` (default: true) | *bool             |          |
| namespace       | The namespace to deploy the GitOps operator into, if empty then it will be deployed cluster-wide into kube-system | string            |          |
| gitUrl          | The URL to git repository to clone                           | string            | Yes      |
| gitBranch       | The git branch to use (default: `master`)                    | string            |          |
| gitPath         | The path with in the git repository to look for YAML in (default: `.`) | string            |          |
| gitPollInterval | The frequency with which to fetch the git repository (default: `5m0s`) | string            |          |
| syncInterval    | The frequency with which to sync the manifests in the repository to the cluster (default: `5m0s`) | string            |          |
| gitKey          | The Kubernetes secret to use for cloning, if it does not exist it will be generated (default: `flux-$name-git-deploy`) | string            |          |
| knownHosts      | The contents of the known_hosts file to mount into Flux and helm-operator | string            |          |
| sshConfig       | The contents of the ~/.ssh/config file to mount into Flux and helm-operator | string            |          |
| fluxVersion     | The version to use for flux (default: 1.9.0 )                | string            |          |
| args            | a map of args to pass to flux without -- prepended. See [fluxd](https://docs.fluxcd.io/en/1.19.0/references/daemon/) for a full list | map[string]string |          |

#### Using an internal git repository

```yaml
gitops:
  - namespace: web-dev
    fluxVersion: 1.20.1.flanksource.1
    gitUrl: ssh://git@bitbucket.localhost.co.za:7999/acme/spec-yaml-dev.git
    gitPath: ./workloads
		gitKey: !!template '{{file.Read "deploy.key" |  base64.Encode}}'
  	sshConfig: ""
    knownHosts: ""
```



#### Enabling image scanning and automatic update

```yaml
gitops:
  - namespace: web-dev
    fluxVersion: 1.20.1.flanksource.1
    gitUrl: ssh://git@bitbucket.localhost.co.za:7999/acme/spec-yaml-dev.gitß
		gitKey: !!template '{{file.Read "deploy.key" |  base64.Encode}}'
    disableScanning: false
    args:
      registry-include-image: "*/web/*"
```

Once Flux is configured for image scanning, deployments need opt-in to updates with the following annotations

    fluxcd.io/automated: "true"
    fluxcd.io/tag.podinfod: semver:~1.3

You can turn off the automation with `fluxcd.io/automated: "false"` or with `fluxcd.io/locked: "true"`.

  Things to notice:

* The annotations are made in `metadata.annotations`, not in `spec.template.metadata`.

* The `fluxcd.io/tag...`. references the container name podinfo, this will change based on your container name. If you have multiple containers you would have multiple lines like that.

* The value for the `fluxcd.io/tag....` annotation should includes the filter pattern type, in this case semver.

  

#### Enabling garbage collection

By default, Flux will not remove resources that were removed in git. To enable garbage collection of deleted resources using a mark and sync approach configure as follows:

```yaml
gitops:
  - namespace: web-dev
    gitUrl: !!template https://{{ getenv "GITHUB_BOT_ACCESS_TOKEN" }}@github.com/acme-port/gitops.git
    gitPath: ./specs/clusters/dev01
    fluxVersion: 1.20.0
    args:
      k8s-default-namespace: web-dev
      sync-garbage-collection: true
      k8s-allow-namespace: web-dev
```




#### Enabling kustomize

Add a `.flux.yml` file in the root of your git repository with the following:
```yaml
version: 1 # must be `1`
patchUpdated:
  generators:
  - command: kustomize build .
```

