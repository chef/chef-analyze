# Subsystem: Command Routing and Top-Level Argument Classification

This document describes the `cmd` subsystem area that handles top-level
command routing and argument classification.

## Purpose

The command routing subsystem is responsible for:

1. Determining whether a call is a top-level help invocation.
2. Determining whether a call targets the top-level `report` command.
3. Preserving stable behavior for root command pre-execution checks.

Primary code references:

- `cmd/root.go`
- `cmd/argv.go`
- `cmd/argv_test.go`
- `cmd/root_test.go`

## Current Behavior Contract

`cmd/argv.go` provides pure helper functions that classify the incoming
argument vector:

1. `isReportCommandArgs(args []string) bool`
2. `isTopLevelHelpArgs(args []string) bool`

`cmd/root.go` wraps these helpers for runtime use against `os.Args`.

Contract rules:

1. Top-level help only: `chef-analyze help`, `chef-analyze -h`, `chef-analyze --help`.
2. Subcommand help does not count as top-level help.
3. `report` detection depends on `args[1] == "report"`.

## Extension Guidance

When introducing new top-level command behavior:

1. Add or adjust classification logic in `cmd/argv.go` first.
2. Keep root wrappers thin in `cmd/root.go`; avoid new parsing branches there.
3. Add table-driven tests in `cmd/argv_test.go` for every new branch.
4. Keep root integration tests in `cmd/root_test.go` focused on root-only concerns.
5. If behavior or user-visible command semantics change, update this document
   and the `## Help Documentation` section in `docs/DESIGN.md` in the same PR.

## Risk Notes

1. Behavioral regression risk: broadening help detection can short-circuit normal
   command execution paths.
2. UX consistency risk: changing classification without updating help docs causes
   mismatches between implementation and contributor expectations.
3. Test drift risk: placing argument parsing tests only in root tests can hide
   parsing regressions behind unrelated setup concerns.

Mitigations:

1. Keep parsing helpers pure and deterministic.
2. Use explicit table-driven tests for all accepted and rejected forms.
3. Include doc updates and risk notes in the same PR as code changes.
