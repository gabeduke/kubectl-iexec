# Security Policy

## Supported versions

Only the latest release receives fixes. Older tags are not patched.

## Reporting a vulnerability

Report privately through GitHub's [security advisory form][advisory]. Please
do not open a public issue for a vulnerability.

[advisory]: https://github.com/PixiBixi/kubectl-iexec/security/advisories/new

Expect an acknowledgement within 7 days. If the report is confirmed, the fix
ships in the next release and the advisory is published once it is available.

This is a personal project maintained on a best-effort basis, with no service
level commitment.

## Scope

In scope: the code in this repository and the release artifacts published from
it (archives, checksums, SBOMs, signatures).

Out of scope: vulnerabilities in upstream dependencies, which belong to their
own maintainers, and issues that require an already-compromised local machine
or an already-compromised cluster.

## Verifying a release

Releases are signed keylessly with cosign and carry a build provenance
attestation:

```sh
cosign verify-blob \
  --certificate-identity-regexp 'https://github.com/PixiBixi/kubectl-iexec/.github/workflows/release.yml@.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  --bundle checksums.txt.sigstore.json \
  checksums.txt

gh attestation verify <archive>.tar.gz --repo PixiBixi/kubectl-iexec
```
