#!/bin/bash

TAG=$(cat /github/workspace/.release)

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

if [ -z "$TAG" ]
then
    echo "Release not found.."
else
    echo "Tag: $TAG"
    git tag $TAG
    git push --tags
fi