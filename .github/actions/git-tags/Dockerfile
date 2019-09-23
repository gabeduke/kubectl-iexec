FROM alpine:latest

RUN apk add --no-cache \
    git curl bash

LABEL "com.github.actions.name"="Git tagger"
LABEL "com.github.actions.description"="Push a tag to git"
LABEL "com.github.actions.icon"="mic"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="http://github.com/gabeduke/kubectl-iexec"
LABEL "homepage"="http://github.com/actions"
LABEL "maintainer"="Octocat <octocat@github.com>"

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]