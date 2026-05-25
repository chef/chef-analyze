# API Subsystem TODO and Edge Sweep

Date: 2026-05-26

This repository currently has no top-level `api/` directory. For this sweep,
`pkg/reporting/` is treated as the API adapter subsystem because it owns
Chef API client setup, API query paths, and capture/report orchestration.

## TODO Sweep

Reviewed TODO entries in `pkg/reporting/**/*.go`:

1. `pkg/reporting/capture.go`
- TODO: package placement for shared capture/common concerns.
- Rationale: Medium-term maintainability improvement; no immediate runtime risk.

2. `pkg/reporting/capture.go`
- TODO: policy group API only supports list + local lookup for group selection.
- Rationale: External API limitation; current lookup-by-map is acceptable and explicit.

3. `pkg/reporting/chef_client.go`
- TODO: upstream URL handling requires appending `/` to avoid malformed requests.
- Rationale: Defensive compatibility shim; keep until upstream behavior is fixed.

4. `pkg/reporting/cookbooks.go`
- TODO: order-of-operations (two network operations to compute totals first).
- Rationale: Performance/latency concern, not correctness regression under normal load.

5. `pkg/reporting/cookbooks.go`
- TODO: add pagination for cookbook usage queries.
- Rationale: Highest scale risk in this set; large node sets may be under-reported
  if paging is not eventually addressed.

6. `pkg/reporting/reporting.go`
- TODO: add debug logging when optional config load path is skipped.
- Rationale: Observability improvement; low functional risk.

## Edge Sweep

Key edge behaviors and why they are acceptable today:

1. Missing node cookbook attribute in capture path.
- Behavior: capture tolerates absent `automatic['cookbooks']` and continues.
- Rationale: Supports never-converged or sparse nodes without hard-failing capture.

2. Missing chef package metadata for Kitchen config.
- Behavior: explicit error when chef package/version metadata is absent.
- Rationale: Fail-fast is correct because kitchen config cannot be produced safely.

3. Policy-group capture path lookup.
- Behavior: lists all groups, then resolves target group from returned map.
- Rationale: Matches current Chef API capability; explicit not-found error is surfaced.

4. Policy/cookbook node usage queries in cookbooks report.
- Behavior: query composes filters and reads `pres.Rows` without pagination.
- Rationale: Correct for moderate result sets; pagination TODO tracks scale boundary.

## Follow-up Priority

1. Add pagination to cookbook usage queries in `pkg/reporting/cookbooks.go`.
2. Revisit two-pass cookbook artifact discovery for latency reduction.
3. Keep upstream URL slash workaround in place until dependency behavior changes.
