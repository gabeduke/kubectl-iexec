# Changelog

Release notes are generated from conventional commits by GoReleaser on each
tagged release, see the GitHub Releases page. This file records the fork's
starting point.

## Unreleased

This fork continues [gabeduke/kubectl-iexec](https://github.com/gabeduke/kubectl-iexec)
after upstream went inactive.

- Merged upstream PR #80 (interfaces, unit tests) and #79 (prompt on the tty,
  fixes #78: no ANSI garbage when stdout is piped)
- Go 1.27, client-go 0.37, all dependencies refreshed, promptui 0.9.0
- Exec over WebSocket with SPDY fallback, raw terminal mode and resize
  handling (upstream #77, #72)
- New `--first` flag (upstream #81)
- Module path is now `github.com/PixiBixi/kubectl-iexec`
- Hardened CI, GoReleaser v2 and Krew publishing to PixiBixi/krew-index
