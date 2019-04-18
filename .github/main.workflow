workflow "Release" {
  resolves = ["goreleaser"]
  on = "release"
}

workflow "Tag" {
  resolves = ["Auto-commit", "push-changelog"]
  on = "push"
}

workflow "Test" {
  resolves = ["codecov"]
  on = "push"
}

action "generate-release-changelog" {
  uses = "docker://ferrarimarco/github-changelog-generator:1.15.0.pre.beta"
  secrets = ["CHANGELOG_GITHUB_TOKEN"]
  env = {
    SRC_PATH = "/github/workspace"
  }
  args = "-u gabeduke -p kubectl-iexec --release-branch develop"
}

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  needs = "generate-release-changelog"
  secrets = [
    "GITHUB_TOKEN",
  ]
  args = "release --release-notes=/github/workspace/CHANGELOG.md"
}

action "is-master" {
  uses = "actions/bin/filter@master"
  args = "branch master"
  secrets = ["GITHUB_TOKEN"]
}

action "tag" {
  uses = "./.github/actions/git-tags"
  needs = "is-master"
  secrets = ["GITHUB_TOKEN"]
}

action "generate-tagged-changelog" {
  uses = "docker://ferrarimarco/github-changelog-generator:1.15.0.pre.beta"
  needs = "tag"
  secrets = ["CHANGELOG_GITHUB_TOKEN"]
  env = {
    SRC_PATH = "/github/workspace"
  }
  args = "-u gabeduke -p kubectl-iexec --release-branch develop"
}

action "push-changelog" {
  uses = "docker://whizark/chandler"
  needs = "generate-tagged-changelog"
  secrets = ["CHANDLER_GITHUB_API_TOKEN"]
  env = {
    CHANDLER_WORKDIR = "/github/workspace"
  }
  args = "push"
}

action "bumpver" {
  uses = "./.github/actions/bumpver"
  needs = "tag"
}

action "Auto-commit" {
  uses = "docker://cdssnc/auto-commit-github-action"
  needs = ["bumpver"]
  args = "This is an auto-commit"
  secrets = ["GITHUB_TOKEN"]
}

action "fmt" {
  uses = "pleo-io/actions/gofmt@master"
  args = "fmt"
}

action "lint" {
  needs = ["fmt"]
  uses = "./.github/actions/golang"
  args = "lint"
}

action "test" {
  needs = ["lint"]
  uses = "./.github/actions/gotest"
}

action "codecov" {
  uses = "pleo-io/actions/codecov@master"
  needs = ["test"]
  secrets = ["CODECOV_TOKEN"]
}
