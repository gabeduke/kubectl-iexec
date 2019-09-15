######################
## RELEASE WORKFLOW ##
######################

workflow "Release" {
  resolves = ["goreleaser"]
  on = "release"
}

## GORELEASER RESOLVES ##

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  needs = "generate-release-changelog"
  secrets = [
    "GITHUB_TOKEN",
  ]
  args = "release --release-notes=/github/workspace/CHANGELOG.md"
}

action "generate-release-changelog" {
  uses = "docker://ferrarimarco/github-changelog-generator:1.15.0.pre.beta"
  secrets = ["CHANGELOG_GITHUB_TOKEN"]
  needs = "created-filter"
  env = {
    SRC_PATH = "/github/workspace"
  }
  args = "-u gabeduke -p kubectl-iexec --release-branch develop"
}

action "created-filter" {
  uses = "actions/bin/filter@master"
  args = "action create"
  needs = ["tag-filter"]
}

action "tag-filter" {
  uses = "actions/bin/filter@master"
  args = "tag"
}


##################
## TAG WORKFLOW ##
##################

workflow "Tag" {
  resolves = ["auto-commit", "push-changelog"]
  on = "push"
}

## AUTO-COMMIT RESOLVES ##

action "auto-commit" {
  uses = "./.github/actions/auto-commit"
  needs = ["bumpver"]
  args = "This is an auto-commit"
  secrets = ["GITHUB_TOKEN"]
}


action "bumpver" {
  uses = "./.github/actions/bumpver"
  needs = "tag"
}

## PUSH-CHANGELOG RESOLVES ##

action "push-changelog" {
  uses = "docker://whizark/chandler"
  needs = "generate-tagged-changelog"
  secrets = ["CHANDLER_GITHUB_API_TOKEN"]
  env = {
    CHANDLER_WORKDIR = "/github/workspace"
  }
  args = "push"
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

## COMMON RESOLVES ##

action "tag" {
  uses = "./.github/actions/git-tags"
  needs = "is-master"
  secrets = ["GITHUB_TOKEN"]
}

action "is-master" {
  uses = "actions/bin/filter@master"
  args = "branch master"
  secrets = ["GITHUB_TOKEN"]
}

##################:
## TEST WORKFLOW ##
##################:

workflow "Test" {
  resolves = ["fmt", "lint", "codecov"]
  on = "pull_request"
}

## FMT RESOLVES ##

action "fmt" {
  uses = "./.github/actions/golang"
  args = "fmt"
  secrets = ["GITHUB_TOKEN"]
}

## LINT RESOLVES ##

action "lint" {
  uses = "./.github/actions/golang"
  args = "lint"
  secrets = ["GITHUB_TOKEN"]
}

## CODECOV RESOLVES ##

action "test" {
  uses = "./.github/actions/golang"
  args = "test"
  secrets = ["GITHUB_TOKEN"]
}

action "codecov" {
  uses = "pleo-io/actions/codecov@master"
  needs = ["test"]
  secrets = ["CODECOV_TOKEN"]
}
