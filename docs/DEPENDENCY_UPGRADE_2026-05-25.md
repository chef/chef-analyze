# Dependency Upgrade Note (2026-05-25)

## Change

- Upgraded `golang.org/x/term` from `v0.40.0` to `v0.41.0` in `go.mod`.
- Indirectly upgraded `golang.org/x/sys` from `v0.41.0` to `v0.42.0`.
- Regenerated vendored dependencies (`go mod vendor`) and module metadata (`go.sum`, `vendor/modules.txt`).

## Why this is low-risk

- `golang.org/x/term` is used by formatter terminal-width handling, not core Chef API/report generation logic.
- Upgrade is a one-step minor bump (`v0.40.0 -> v0.41.0`) to reduce drift while limiting change surface.

## Validation Performed

Executed in a Dockerized Go 1.26 environment:

```bash
docker run --rm -v "$PWD":/src -w /src golang:1.26 \
  bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && GOFLAGS=-mod=mod go test ./pkg/formatter"
```

Result:

- `ok github.com/chef/chef-analyze/pkg/formatter`

## Impact

- No functional code changes in command/reporting flows.
- Dependency/runtime behavior changes are limited to terminal capability helpers from `x/term` and transitive `x/sys` internals.
- Vendored source updates are expected for reproducible builds.

## Rollback

If any issue appears after merge, rollback options:

1. Revert commit that introduced this upgrade.
2. Manually pin previous versions and re-vendor:

```bash
go get golang.org/x/term@v0.40.0 golang.org/x/sys@v0.41.0
go mod tidy
go mod vendor
```

3. Re-run formatter tests to confirm rollback state:

```bash
go test ./pkg/formatter
```
