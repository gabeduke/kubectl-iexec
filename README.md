# kubectl-iexec

[![CI](https://github.com/PixiBixi/kubectl-iexec/actions/workflows/ci.yml/badge.svg)](https://github.com/PixiBixi/kubectl-iexec/actions/workflows/ci.yml)
[![Lint](https://github.com/PixiBixi/kubectl-iexec/actions/workflows/lint.yml/badge.svg)](https://github.com/PixiBixi/kubectl-iexec/actions/workflows/lint.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Interactive `kubectl exec`: filter pods by name, pick a pod and a container from a list, and get a shell (or run a command) in it.

This is a maintained fork of [gabeduke/kubectl-iexec](https://github.com/gabeduke/kubectl-iexec), which has been inactive since 2023. All credit for the original design and code goes to gabeduke. The fork keeps the MIT license and the original copyright.

## What changed compared to upstream

- Go 1.27, client-go 0.37, current dependencies
- WebSocket exec with SPDY fallback (like modern `kubectl`), raw terminal mode and live resize
- Clean output when stdin or stdout is piped (`kubectl iexec api printenv | sort > .env`)
- `--first` to skip the prompts
- Releases signed with cosign, with SBOMs and provenance, published to a Krew index

## Install

With [Krew](https://krew.sigs.k8s.io/):

```bash
kubectl krew index add pixibixi https://github.com/PixiBixi/krew-index
kubectl krew install pixibixi/iexec
```

With Go:

```bash
go install github.com/PixiBixi/kubectl-iexec/v2@latest
```

The binary must be named `kubectl-iexec` and be on your `PATH`. Prebuilt archives for Linux and macOS (amd64, arm64) are on the [releases page](https://github.com/PixiBixi/kubectl-iexec/releases).

## Usage

```text
kubectl iexec [pod filter] [remote command(s)] [flags]
```

The first argument filters pods by name. The remaining arguments are the command to run (default: `bash` when the image has it, `sh` otherwise).

```bash
# pick from all pods in the namespace, then open /bin/sh
kubectl iexec

# pick from pods matching "busybox"
kubectl iexec busybox

# run a command
kubectl iexec busybox cat /etc/hosts

# pick a pod, then a container matching "second_container"
kubectl iexec multi-container-pod -c second_container

# no prompt: first matching pod and container
kubectl iexec busybox --first cat /etc/hosts
```

| Flag | Description |
| --- | --- |
| `-c, --container` | Container name filter |
| `-l, --selector` | Label selector to filter pods (`key1=value1,key2=value2`) |
| `--first` | Skip the prompts and use the first matching pod and container |
| `--all-phases` | Include pods that are not `Running` or are terminating (skipped by default) |
| `-A, --all-namespaces` | Search pods across all namespaces |
| `-n, --namespace` | Namespace to search |
| `-v, --vim-mode` | Vim keys in the prompts |
| `-x, --naked` | Decolorize output |
| `--log-level` | `trace`, `debug`, `info`, `warn`, `error`, `fatal` or `panic` |

The usual kubectl connection flags (`--context`, `--kubeconfig`, `--server`, ...) are also accepted.

## Development

```bash
make build   # build ./kubectl-iexec
make test    # go test -race ./...
make lint    # golangci-lint
```

See [AGENTS.md](AGENTS.md) for the full workflow and [SECURITY.md](SECURITY.md) to report a vulnerability.

## License

MIT, see [LICENSE](LICENSE). Original copyright notice kept.
