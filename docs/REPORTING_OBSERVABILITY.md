# Reporting Observability Pattern

This document defines the logging and metrics pattern used across
`pkg/reporting/`.

## Pattern

`pkg/reporting` uses an optional package-level sink with no-op default:

- `SetObservabilitySink(sink ObservabilitySink)`
- `emitLog(event, keyValues...)`
- `emitMetric(name, value, keyValues...)`

Code references:

- `pkg/reporting/observability.go`
- `pkg/reporting/reporting.go`
- `pkg/reporting/capture.go`
- `pkg/reporting/cookbooks.go`
- `pkg/reporting/nodes.go`

## Event Naming Convention

Use dotted names scoped by subsystem and action:

- `reporting.<area>.<action>.<state>`

Examples:

- `reporting.capture.run.start`
- `reporting.cookbooks.generate.duration_ms`
- `reporting.nodes.generate.result_count`

## Validation

1. Unit test the sink wiring:

```bash
go test ./pkg/reporting -run 'TestSetObservabilitySink'
```

2. Run package tests for reporting:

```bash
go test ./pkg/reporting
```

3. Run cmd tests to ensure integration callers are unaffected:

```bash
go test ./cmd
```

## Extension Guidance

When adding new reporting flows:

1. Emit one start metric/log and one completion metric/log.
2. Emit duration in milliseconds for long-running operations.
3. Emit an explicit error metric/log in each failure return path.
4. Keep key/value labels concise and deterministic.
