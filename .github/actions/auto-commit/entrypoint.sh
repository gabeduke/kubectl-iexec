#!/bin/bash
set -e

BRANCH=develop

git config --global user.name 'autobot'
git config --global user.email 'autobot@leetserve.com'

git add .release
git fetch origin
git checkout "${BRANCH}"

git commit -m  "bumpver" --allow-empty
git push -u origin "${BRANCH}"