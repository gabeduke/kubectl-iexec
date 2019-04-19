#!/bin/sh
set -e

BRANCH=patch-${GITHUB_SHA}

git remote set-url origin https://gabeduke:${GITHUB_TOKEN}@github.com/gabeduke/kubectl-iexec.git
git config --global user.name 'autobot'
git config --global user.email 'autobot@leetserve.com'
git checkout -b "${BRANCH}"
git add -A && git commit -m '$*' --allow-empty
git push -u origin "${BRANCH}"

curl -v \
    -u gabeduke:$GITHUB_TOKEN \
    -H "Content-Type:application/json" \
    -X POST https://api.github.com/repos/gabeduke/kubectl-iexec/pulls \
    -d '{"title":"bumpver", "body": "bumpity bump", "head": "'"${BRANCH}"'", "base": "develop"}'
