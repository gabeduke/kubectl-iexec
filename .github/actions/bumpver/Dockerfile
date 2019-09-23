FROM alpine:latest

RUN apk add --no-cache \
    bash

LABEL "com.github.actions.name"="bumpver"
LABEL "com.github.actions.description"="Push a tag to git"
LABEL "com.github.actions.icon"="github"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="http://github.com/gabeduke/kubectl-iexec"
LABEL "homepage"="http://github.com/actions"
LABEL "maintainer"="Octocat <octocat@github.com>"

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]