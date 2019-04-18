#!/bin/sh -l

TAG=$(cat /github/workspace/.release)

if [ -z "$TAG" ]
then
    echo "Release not found.."
else
    echo "Tag: $TAG"
    git tag $TAG
    git push origin $TAG
fi
