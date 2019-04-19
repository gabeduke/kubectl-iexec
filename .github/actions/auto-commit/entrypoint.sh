#!/bin/bash
set -e

BRANCH=patch-$(git rev-parse --short HEAD)

git config --global user.name 'autobot'
git config --global user.email 'autobot@leetserve.com'
git checkout -b "${BRANCH}"
git add -A && git commit -m 'auto-commit' --allow-empty
git push -u origin "${BRANCH}"

curl -v \
    -u gabeduke:$GITHUB_TOKEN \
    -H "Content-Type:application/json" \
    -H "Accept: application/vnd.github.v3+json" \
    -X POST https://api.github.com/repos/gabeduke/kubectl-iexec/pulls \
    -d '{"title":"bumpver", "body": "bumpity bump", "head": "'"${BRANCH}"'", "base": "develop"}'
