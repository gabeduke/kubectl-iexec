#!/bin/sh
set -e

BRANCH=patch-${GITHUB_SHA}

git config --global user.name 'autobot'
git config --global user.email 'autobot@leetserve.com'
git checkout -b "${BRANCH}"
git add -A && git commit -m '$*' --allow-empty
git push -u origin "${BRANCH}"