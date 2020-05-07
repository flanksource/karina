![](https://docs.fluxcd.io/en/1.19.0/_files/flux-cd-diagram.png)


#### Using annotations

    fluxcd.io/automated: "true"
    fluxcd.io/tag.podinfod: semver:~1.3

You can turn off the automation with `fluxcd.io/automated: "false"` or with `fluxcd.io/locked: "true"`.

  Things to notice:

* The annotations are made in metadata.annotations, not in spec.template.metadata.
* The fluxcd.io/tag.... references the container name podinfod, this will change based on your container name. If you have multiple containers you would have multiple lines like that.
* The value for the fluxcd.io/tag.... annotation should includes the filter pattern type, in this case semver.


#### Using kustomize

Add a `.flux.yml` file in the root of your repository:
```yaml
version: 1 # must be `1`
patchUpdated:
  generators:
  - command: kustomize build .
```
