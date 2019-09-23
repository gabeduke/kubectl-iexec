#!/bin/bash
set -e

BRANCH=develop

cat <<- EOF > $HOME/.netrc
        machine github.com
        login $GITHUB_ACTOR
        password $GITHUB_TOKEN
        machine api.github.com
        login $GITHUB_ACTOR
        password $GITHUB_TOKEN
EOF

chmod 600 $HOME/.netrc

git config --global user.name 'autobot'
git config --global user.email 'autobot@leetserve.com'

git add .release
git fetch origin
git checkout "${BRANCH}"

git commit -m  "bumpver" --allow-empty
git push -u origin "${BRANCH}"