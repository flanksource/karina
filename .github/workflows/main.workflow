workflow "Release" {
  on = "push"
  resolves = ["goreleaser"]
}

action "is-tag" {
  uses = "actions/bin/filter@master"
  args = "tag"
}

action "goreleaser" {
  uses = "goreleaser/goreleaser/action@actions"
  secrets = [
    "GITHUB_TOKEN"
  ]
  args = "release"
  needs = ["is-tag"]
}
