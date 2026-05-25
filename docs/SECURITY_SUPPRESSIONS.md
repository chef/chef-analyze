# Security Suppressions

This document tracks intentional static-analysis suppressions with rationale.

## Scope

- Scanner: `gosec`
- Strict gated path: `cmd/**`
- Strict rules: `G104`, `G115`, `G204`, `G301`, `G304`

## Active suppressions

1. `cmd/s3_utils.go` `UploadToS3` `os.Open(filePath)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: `filePath` is an explicit user-provided CLI argument for upload behavior; opening this path is required to read local content for S3 upload.

2. `cmd/s3_utils.go` `writePrivateFile` `os.OpenFile(filePath, ...)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: path is derived from `ChefWorkstationDir()` and generated filename segments (`token-<minutes>-<timestamp>`), not raw external input; writes are centralized and forced to file mode `0600`.

3. `cmd/capture.go` `exec.Command("chef", "generate", "repo", repoName)`
- Rule: `G204`
- Suppression: `#nosec G204`
- Rationale: `repoName` is derived from `nodeName` only after strict allowlist validation (`^[A-Za-z0-9._-]+$`), which constrains subprocess arguments to safe characters.

4. `cmd/report.go` `writeReportFile` `os.OpenFile(reportPath, ...)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: report paths are built from `ChefWorkstationDir()` and controlled filename segments (`timestamp`, fixed extensions, optional `-filtered` suffix), and writes enforce `0600` mode.

## Remediations paired with strict scan

- `G301` remediated in `cmd/s3_utils.go`: token directory permissions changed from `os.ModePerm` to `0o700`.
- `G301` remediated in `cmd/report.go`: output directories now use `0o750`.
- `G115` remediated previously: session duration conversion now validates bounds before converting to `int32`.
- Nil-pointer hardening in `cmd/s3_utils.go`: credential renderers now guard missing nested fields before dereference.
- File-write hygiene in `cmd/s3_utils.go`: token payload and shell exports are written through one helper that checks write and close errors and enforces `0600` permissions.
- `G104` remediated in `cmd/`: previously ignored errors are now checked in `root`, `capture`, and report file writes.

## Review guidance

- Revisit suppressions when path handling is refactored.
- Prefer replacing suppressions with stronger path-scoping APIs when practical.

## Automated patch set

Use the scripted patch set helper for repeatable security maintenance:

```bash
bash scripts/security_alert_patchset.sh --apply-minor
```

Optional PR creation (requires authenticated `gh` CLI):

```bash
bash scripts/security_alert_patchset.sh --apply-minor --create-pr --pr-title "Security hygiene patch set"
```
