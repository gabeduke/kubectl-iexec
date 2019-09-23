FROM alpine:latest

LABEL "name"="codecov"
LABEL "maintainer"="Gabe Duke <gabeduke@gmail.com>"
LABEL "version"="1.0.0"

LABEL "com.github.actions.name"="Codecov for GitHub Actions"
LABEL "com.github.actions.description"="Submit code coverage"
LABEL "com.github.actions.icon"="git"
LABEL "com.github.actions.color"="orange"

RUN apk add --no-cache curl bash git

ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]