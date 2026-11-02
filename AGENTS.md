# AGENTS.md

Guidance for coding agents working in this repository.

## What this is

`kubectl-iexec` is a kubectl plugin (`kubectl iexec`): it filters pods, lets
the user pick a pod and a container (promptui), then execs into it. Fork of
gabeduke/kubectl-iexec. Go 1.27, cobra, client-go (typed client and
`remotecommand`), promptui, `golang.org/x/term`. The module path is
`github.com/PixiBixi/kubectl-iexec/v2`.

## Common commands

```bash
make build      # go build -ldflags "-s" -o kubectl-iexec .
make test       # go test -race ./...
make lint       # golangci-lint run (config: .golangci.yml, version "2")
make lint-md FILES="README.md AGENTS.md"  # markdownlint, as the CI job runs it
make snapshot   # goreleaser release --snapshot --clean (dry-run)

go test -race ./pkg/iexec -run TestIexec_Do_First   # single test
pre-commit run --all-files
```

CI lives in `.github/workflows/`: `ci.yml` (verify, build, test -race),
`lint.yml`, `go-format`, `govulncheck`, `codeql`, `dependency-review`,
`scorecard`, `github-actions` (zizmor), `markdownlint`, `validate-pr-title`
and `release.yml`. Validate workflows with `actionlint` (runner labels are
declared in `.github/actionlint.yaml`).

## Architecture

- `kubectl-iexec.go` / `cmd/root.go`: entry point, wires cobra to stdio.
- `pkg/cmd/iexec.go`: flags (`IExecOptions`), builds `iexec.Config` and the
  real implementations.
- `pkg/iexec/iexec.go`: `Iexec.Do` orchestrates fetch, match, select, exec.
  The three collaborators are interfaces in `iface.go`, so tests inject mocks
  (`iexec_test.go`): `TerminalUI` (`terminal_ui_prompt.go`, promptui on
  `/dev/tty` so piped stdout stays clean), `K8sClient`
  (`k8s_client_impl.go`) and `ExecRunner` (`exec_runner.go`).
- `exec_runner.go`: WebSocket executor with SPDY fallback, remote TTY and
  raw local terminal only when stdin is a terminal; `resize_unix.go` forwards
  SIGWINCH, `resize_windows.go` sends the initial size only.

Adding a flag: declare it in `pkg/cmd/iexec.go`, add a field to
`iexec.Config`, use it in `Iexec.Do`, add a test, update the README flag table.

## Testing

Unit tests use hand-written mocks of the three interfaces. The exec path
itself needs a cluster and is not covered by unit tests; test it by hand
against a real pod (including `| cat` piping and a window resize).

## Releasing

Releases are automatic on push to `main`: `release.yml` runs `svu` to compute
the next `vX.Y.Z` from conventional commits (`feat` minor, `fix` patch,
anything else no release), tags, then runs GoReleaser. GoReleaser publishes
archives, SBOMs, a cosign signature and the Krew manifest `plugins/iexec.yaml`
to PixiBixi/krew-index (needs the `KREW_INDEX_TOKEN` secret). Renovate drives
dependency bumps (`renovate.json`).
