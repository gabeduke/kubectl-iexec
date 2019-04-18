FROM golangci/golangci-lint:v1.15

LABEL "com.github.actions.name"="Lint"
LABEL "com.github.actions.description"="Lint go gocode"
LABEL "com.github.actions.icon"="mic"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="http://github.com/gabeduke/kubectl-iexec"
LABEL "homepage"="http://github.com/actions"
LABEL "maintainer"="Octocat <octocat@github.com>"

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]