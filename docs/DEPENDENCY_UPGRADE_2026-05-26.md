# Dependency Upgrade Note (2026-05-26)

## Change Set

Minor dependency upgrades applied:

1. `github.com/aws/aws-sdk-go-v2/config`:
- `v1.29.14` -> `v1.32.18`

2. `github.com/aws/aws-sdk-go-v2/service/sts`:
- `v1.33.19` -> `v1.42.1`

3. `golang.org/x/term`:
- `v0.41.0` -> `v0.43.0`

Transitive updates were refreshed by design (for example `aws-sdk-go-v2`,
`aws/smithy-go`, `golang.org/x/sys`) and vendored sources were regenerated.

## Sweep Scope

Minor upgrade sweep was applied across vendored dependency sources:

- `vendor/github.com/aws/**`
- `vendor/golang.org/x/**`
- `vendor/modules.txt`

This keeps the repository's vendored dependency folder aligned with `go.mod`
and `go.sum`.

## Why This Is Safe

1. All upgrades are minor-version bumps on existing dependencies already used
   by the project.
2. No application logic changes were introduced in `cmd/` or `pkg/` runtime code.
3. Validation was run against CI-equivalent checks used in this repository.

## Validation (CI-Equivalent)

Executed locally using the same commands/patterns as workflow jobs:

1. cmd tests with coverage:

```bash
GO_BIN="$(brew --prefix go)/libexec/bin/go"
mkdir -p coverage
"$GO_BIN" test ./cmd -coverprofile=coverage/cmd.out -covermode=atomic
"$GO_BIN" tool cover -func=coverage/cmd.out | grep '^total:'
```

Observed:

- `ok github.com/chef/chef-analyze/cmd`
- `total: (statements) 23.0%`

2. formatter contract tests:

```bash
"$GO_BIN" test ./pkg/formatter -run '^TestContract_'
```

Observed:

- `ok github.com/chef/chef-analyze/pkg/formatter`

3. strict gosec scan parity with workflow targeting `cmd/s3_utils.go`:

```bash
"$GO_BIN" install github.com/securego/gosec/v2/cmd/gosec@latest
"$($GO_BIN env GOPATH)/bin/gosec" -fmt=json -no-fail -include=G115,G301,G304 ./cmd/... > /tmp/gosec.json
```

Observed:

- No strict-path gosec issues in `cmd/s3_utils.go`

## Rollback

If regressions appear after merge, use either full revert or targeted re-pin.

Option 1: Revert the upgrade commit.

Option 2: Re-pin upgraded dependencies and re-vendor.

```bash
GO_BIN="$(brew --prefix go)/libexec/bin/go"
GOFLAGS=-mod=mod "$GO_BIN" get \
  github.com/aws/aws-sdk-go-v2/config@v1.29.14 \
  github.com/aws/aws-sdk-go-v2/service/sts@v1.33.19 \
  golang.org/x/term@v0.41.0
GOFLAGS=-mod=mod "$GO_BIN" mod tidy
GOFLAGS=-mod=mod "$GO_BIN" mod vendor
```

Then rerun the same CI-equivalent validation commands listed above.
