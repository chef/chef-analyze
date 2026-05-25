# Security Suppressions

This document tracks intentional static-analysis suppressions with rationale.

## Scope

- Scanner: `gosec`
- Strict gated path: `cmd/s3_utils.go`
- Strict rules: `G115`, `G301`, `G304`

## Active suppressions

1. `cmd/s3_utils.go` `UploadToS3` `os.Open(filePath)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: `filePath` is an explicit user-provided CLI argument for upload behavior; opening this path is required to read local content for S3 upload.

2. `cmd/s3_utils.go` `saveSessionToken` `os.Create(sessionToken)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: path is derived from `ChefWorkstationDir()` and generated filename segments (`token-<minutes>-<timestamp>`), not raw external input.

3. `cmd/s3_utils.go` `saveSessionToken` `os.Create(shFileName)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: path is generated from trusted base directory and controlled filename formatting.

4. `cmd/s3_utils.go` `saveSessionToken` `os.Create(ps1FileName)`
- Rule: `G304`
- Suppression: `#nosec G304`
- Rationale: path is generated from trusted base directory and controlled filename formatting.

## Remediations paired with strict scan

- `G301` remediated in `cmd/s3_utils.go`: token directory permissions changed from `os.ModePerm` to `0o700`.
- `G115` remediated previously: session duration conversion now validates bounds before converting to `int32`.

## Review guidance

- Revisit suppressions when path handling is refactored.
- Prefer replacing suppressions with stronger path-scoping APIs when practical.
