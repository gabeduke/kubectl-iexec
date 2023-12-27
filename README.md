![](https://github.com/gabeduke/kubectl-iexec/workflows/Test/badge.svg)
![](https://github.com/gabeduke/kubectl-iexec/workflows/Lint/badge.svg)
![](https://github.com/gabeduke/kubectl-iexec/workflows/Fmt/badge.svg)

# Kubectl Interactive Exec

## Summary

`Kubectl-iexec` is a plugin providing an interactive selector to exec into a running pod. For a search filter, the plugin will return a list of pods and containers that match, then perform a `kubectl exec` to the selection.

_Notes:_

Kubectl >= `v1.12.0` is required for plugins to work

For more information on kuberctl plugins see [documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)


## Usage:

```
$ kubectl iexec --help

Kubectl-iexec is an interactive pod and container selector for `kubectl exec`

Arg[1] will act as a filter, any pods that match will be returned in a list
that the user can select from. Subsequent args make up the array of commands that
should be executed on the pod.

example:
kubectl iexec busybox /bin/sh

example:
kubectl iexec busybox cat /etc/hosts

Usage:
  iexec [pod filter] [remote command(s)] [flags]

Flags:
  -c, --container string   Container to search
  -h, --help               help for iexec
  --log-level string   log level (trace|debug|info|warn|error|fatal|panic)
  -x, --naked              Decolorize output
  -n, --namespace string   Namespace to search
  -v, --vim-mode            Vim Mode enabled
  -l, --label string        Label selector to filter pods
```


[![asciicast](https://asciinema.org/a/kW4u4sPSaFo77O7BaSwMM1Kuh.svg)](https://asciinema.org/a/kW4u4sPSaFo77O7BaSwMM1Kuh)


## Install:

To install the plugin, the binary simply needs to be somewhere on your path in the format kubectl-[plugin_name]. the simplest way to do this is to `go get` the package:

`go get -u github.com/gabeduke/kubectl-iexec`

Alternatively you may pull the binary from the releases page on Github:

Select OS
```bash
# Linux
OS=LINUX

# Mac
OS=DARWIN
```

Run:
```bash
# Get latest release
TAG=$(curl -s https://api.github.com/repos/gabeduke/kubectl-iexec/releases/latest | grep -oP '"tag_name": "\K(.*)(?=")')

# Donwload and extract binary to /usr/local/bin
curl -LO https://github.com/gabeduke/kubectl-iexec/releases/download/${TAG}/kubectl-iexec_${TAG}_${OS:-Linux}_x86_64.tar.gz

mkdir -p /tmp/kubectl-iexec
tar -xzvf kubectl-iexec_${TAG}_${OS}_x86_64.tar.gz -C /tmp/kubectl-iexec
chmod +x /tmp/kubectl-iexec/kubectl-iexec

sudo mv /tmp/kubectl-iexec/kubectl-iexec /usr/local/bin
```
